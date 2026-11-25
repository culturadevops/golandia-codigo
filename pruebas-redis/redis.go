// redis.go — un servidor Redis escrito en Go, para el video 198 de Golandia.
//
// LA PROMESA DEL VIDEO, que este fichero tiene que cumplir: que el `redis-cli` OFICIAL se
// conecte a esto y funcione sin notar la diferencia. Por eso habla RESP de verdad y contesta
// a COMMAND DOCS, que es lo primero que manda redis-cli al conectarse (si no, se queda colgado).
//
// EL GIRO DEL VIDEO: Redis es de UN SOLO HILO, y es rapido POR ESO. Aqui estan las dos
// implementaciones para poder medirlo:
//
//	-modo=mutex   un mapa compartido y un candado por operacion (lo que sale natural en Go)
//	-modo=hilo    UN solo goroutine dueno del mapa; las ordenes le llegan por un canal
//
// LO QUE NO HACE, y en el video se dice claro: no hay persistencia (RDB/AOF), ni replicacion,
// ni cluster, ni transacciones, ni pub/sub. Eso es lo que hace que Redis sea Redis y no un mapa
// con un puerto.
//
//	go run redis.go -puerto 6380 -modo hilo
package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"log"
	"math"
	"net"
	"strconv"
	"strings"
	"sync"
	"time"
)

// ─────────────────────────────────────────────────────────── el almacen ──

type entrada struct {
	valor string
	// cero = no caduca nunca. Se guarda el INSTANTE, no la duracion: si guardas la duracion
	// tienes que ir restando, y eso es un bucle que toca todas las claves.
	caduca time.Time
}

type almacen struct {
	mu    sync.Mutex
	datos map[string]entrada
	hilo  bool // true = un solo goroutine manda; el candado no se usa
}

func nuevo(hilo bool) *almacen {
	return &almacen{datos: map[string]entrada{}, hilo: hilo}
}

// vivo dice si la entrada sigue valida, y BORRA la caducada de paso.
//
// Esto es el "borrado perezoso" de Redis: no hay ningun proceso mirando el reloj de cada clave;
// la clave muerta se descubre cuando alguien la pide. Lo que hace el barrido periodico de abajo
// es limpiar lo que nadie vuelve a pedir, que si no crece para siempre.
func (a *almacen) vivo(k string) (entrada, bool) {
	e, hay := a.datos[k]
	if !hay {
		return entrada{}, false
	}
	if !e.caduca.IsZero() && time.Now().After(e.caduca) {
		delete(a.datos, k)
		return entrada{}, false
	}
	return e, true
}

// barrer: el otro medio de la expiracion. Redis muestrea claves con caducidad al azar; aqui,
// como el objetivo es ensenarlo y no batir marcas, se recorre entero cada segundo.
func (a *almacen) barrer(hecho <-chan struct{}) {
	t := time.NewTicker(time.Second)
	defer t.Stop()
	for {
		select {
		case <-hecho:
			return
		case <-t.C:
			a.conCandado(func() {
				ahora := time.Now()
				for k, e := range a.datos {
					if !e.caduca.IsZero() && ahora.After(e.caduca) {
						delete(a.datos, k)
					}
				}
			})
		}
	}
}

// conCandado ejecuta f protegiendo el mapa SOLO en modo mutex. En modo hilo no hace falta
// candado porque solo hay un goroutine tocando el mapa: esa es toda la tesis del video.
// ⚠️ EL QUE PIDE TIENE QUE ESPERAR. Primera version de esto: `peticiones <- f` y a correr. Con un
// canal sin buffer el envio termina cuando el dueno RECIBE la funcion, no cuando la EJECUTA, asi
// que `ejecutar` leia su variable de respuesta antes de que nadie la escribiera y contestaba la
// cadena vacia. El redis-cli oficial se quedaba colgado esperando un RESP que nunca llegaba
// entero: ni error, ni traza, ni nada en los registros. Encontrado el 2026-09-05.
func (a *almacen) conCandado(f func()) {
	if a.hilo {
		listo := make(chan struct{})
		peticiones <- func() { f(); close(listo) }
		<-listo
		return
	}
	a.mu.Lock()
	f()
	a.mu.Unlock()
}

// El canal por el que viajan las ordenes en modo hilo. Sin buffer a proposito: el que pide
// espera a que el dueno del mapa termine, que es justo lo que hace Redis.
var peticiones = make(chan func())

func unicoHilo(hecho <-chan struct{}) {
	for {
		select {
		case <-hecho:
			return
		case f := <-peticiones:
			f()
		}
	}
}

// ───────────────────────────────────────────────────────────── comandos ──

func (a *almacen) ejecutar(partes []string) string {
	if len(partes) == 0 {
		return "-ERR peticion vacia\r\n"
	}
	orden := strings.ToUpper(partes[0])
	args := partes[1:]

	switch orden {

	case "PING":
		if len(args) == 1 {
			return bulk(args[0])
		}
		return "+PONG\r\n"

	case "ECHO":
		if len(args) != 1 {
			return errArgs(orden)
		}
		return bulk(args[0])

	// redis-cli manda COMMAND DOCS nada mas conectarse. Si no contestas algo, se queda
	// esperando y parece que el servidor esta roto. Un array vacio le vale.
	case "COMMAND":
		return "*0\r\n"

	case "SET":
		if len(args) < 2 {
			return errArgs(orden)
		}
		var caduca time.Time
		// SET k v EX <segundos> / PX <milisegundos>
		for i := 2; i+1 < len(args); i += 2 {
			n, err := strconv.Atoi(args[i+1])
			if err != nil {
				return "-ERR value is not an integer or out of range\r\n"
			}
			switch strings.ToUpper(args[i]) {
			case "EX":
				caduca = time.Now().Add(time.Duration(n) * time.Second)
			case "PX":
				caduca = time.Now().Add(time.Duration(n) * time.Millisecond)
			}
		}
		a.conCandado(func() { a.datos[args[0]] = entrada{valor: args[1], caduca: caduca} })
		return "+OK\r\n"

	case "GET":
		if len(args) != 1 {
			return errArgs(orden)
		}
		var r string
		a.conCandado(func() {
			if e, hay := a.vivo(args[0]); hay {
				r = bulk(e.valor)
			} else {
				r = "$-1\r\n" // nil: la clave no esta, o caduco
			}
		})
		return r

	case "DEL":
		var n int
		a.conCandado(func() {
			for _, k := range args {
				if _, hay := a.vivo(k); hay {
					delete(a.datos, k)
					n++
				}
			}
		})
		return fmt.Sprintf(":%d\r\n", n)

	case "EXISTS":
		var n int
		a.conCandado(func() {
			for _, k := range args {
				if _, hay := a.vivo(k); hay {
					n++
				}
			}
		})
		return fmt.Sprintf(":%d\r\n", n)

	case "EXPIRE":
		if len(args) != 2 {
			return errArgs(orden)
		}
		seg, err := strconv.Atoi(args[1])
		if err != nil {
			return "-ERR value is not an integer or out of range\r\n"
		}
		var n int
		a.conCandado(func() {
			if e, hay := a.vivo(args[0]); hay {
				e.caduca = time.Now().Add(time.Duration(seg) * time.Second)
				a.datos[args[0]] = e
				n = 1
			}
		})
		return fmt.Sprintf(":%d\r\n", n)

	case "TTL":
		if len(args) != 1 {
			return errArgs(orden)
		}
		var r int64
		a.conCandado(func() {
			e, hay := a.vivo(args[0])
			switch {
			case !hay:
				r = -2 // no existe
			case e.caduca.IsZero():
				r = -1 // existe y no caduca
			default:
				// Hacia ARRIBA, como el Redis de verdad: recien puesto un EX 2, `TTL` contesta
				// 2, no 1. Truncando salia 1 y el redis-cli oficial enseñaba la diferencia.
				r = int64(math.Ceil(time.Until(e.caduca).Seconds()))
			}
		})
		return fmt.Sprintf(":%d\r\n", r)

	case "INCR", "DECR":
		if len(args) != 1 {
			return errArgs(orden)
		}
		paso := 1
		if orden == "DECR" {
			paso = -1
		}
		var r string
		a.conCandado(func() {
			n := 0
			if e, hay := a.vivo(args[0]); hay {
				v, err := strconv.Atoi(e.valor)
				if err != nil {
					r = "-ERR value is not an integer or out of range\r\n"
					return
				}
				n = v
			}
			n += paso
			// Se conserva la caducidad: INCR no reinicia el reloj de la clave.
			vieja := a.datos[args[0]]
			a.datos[args[0]] = entrada{valor: strconv.Itoa(n), caduca: vieja.caduca}
			r = fmt.Sprintf(":%d\r\n", n)
		})
		return r

	case "DBSIZE":
		var n int
		a.conCandado(func() {
			for k := range a.datos {
				if _, hay := a.vivo(k); hay {
					n++
				}
			}
		})
		return fmt.Sprintf(":%d\r\n", n)

	case "FLUSHALL", "FLUSHDB":
		a.conCandado(func() { a.datos = map[string]entrada{} })
		return "+OK\r\n"

	case "QUIT":
		return "+OK\r\n"

	default:
		return fmt.Sprintf("-ERR unknown command '%s'\r\n", partes[0])
	}
}

func bulk(s string) string { return fmt.Sprintf("$%d\r\n%s\r\n", len(s), s) }

func errArgs(o string) string {
	return fmt.Sprintf("-ERR wrong number of arguments for '%s' command\r\n", strings.ToLower(o))
}

// ───────────────────────────────────────────────────────────────── RESP ──

// leerOrden lee UNA orden del cliente.
//
// RESP es TEXTO, y ahi esta media gracia del video: una orden es un array de cadenas con la
// longitud por delante.
//
//	*3\r\n $3\r\nSET\r\n $3\r\nfoo\r\n $3\r\nbar\r\n
//
// Tambien se acepta la forma "inline" (una linea suelta con espacios) porque es lo que manda
// telnet o netcat, y en el video se ensena conectando con netcat.
func leerOrden(r *bufio.Reader) ([]string, error) {
	linea, err := r.ReadString('\n')
	if err != nil {
		return nil, err
	}
	linea = strings.TrimRight(linea, "\r\n")
	if linea == "" {
		return nil, nil
	}
	if linea[0] != '*' {
		return strings.Fields(linea), nil // inline: netcat, telnet
	}
	n, err := strconv.Atoi(linea[1:])
	if err != nil || n < 0 {
		return nil, fmt.Errorf("array mal formado: %q", linea)
	}
	partes := make([]string, 0, n)
	for i := 0; i < n; i++ {
		cab, err := r.ReadString('\n') // $<longitud>
		if err != nil {
			return nil, err
		}
		cab = strings.TrimRight(cab, "\r\n")
		if len(cab) == 0 || cab[0] != '$' {
			return nil, fmt.Errorf("bulk mal formado: %q", cab)
		}
		largo, err := strconv.Atoi(cab[1:])
		if err != nil {
			return nil, err
		}
		buf := make([]byte, largo+2) // +2 por el \r\n del final
		if _, err := io.ReadFull(r, buf); err != nil {
			return nil, err
		}
		partes = append(partes, string(buf[:largo]))
	}
	return partes, nil
}

func atender(c net.Conn, a *almacen) {
	defer c.Close()
	r := bufio.NewReader(c)
	w := bufio.NewWriter(c)
	for {
		partes, err := leerOrden(r)
		if err != nil {
			return // el cliente se fue
		}
		if partes == nil {
			continue
		}
		w.WriteString(a.ejecutar(partes))
		if err := w.Flush(); err != nil {
			return
		}
		if len(partes) > 0 && strings.ToUpper(partes[0]) == "QUIT" {
			return
		}
	}
}

func main() {
	puerto := flag.String("puerto", "6380", "puerto donde escuchar")
	modo := flag.String("modo", "hilo", "hilo = un solo goroutine dueno del mapa · mutex = candado por operacion")
	// Por defecto SOLO local. Se abre a 0.0.0.0 unicamente para la prueba con el redis-cli
	// oficial, que corre dentro de un contenedor y entra por host.docker.internal.
	bind := flag.String("bind", "127.0.0.1", "interfaz donde escuchar (0.0.0.0 para el contenedor)")
	flag.Parse()

	if *modo != "hilo" && *modo != "mutex" {
		log.Fatalf("modo desconocido: %q (hilo | mutex)", *modo)
	}
	a := nuevo(*modo == "hilo")
	hecho := make(chan struct{})
	if a.hilo {
		go unicoHilo(hecho)
	}
	go a.barrer(hecho)

	ln, err := net.Listen("tcp", *bind+":"+*puerto)
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("escuchando en %s:%s  ·  modo=%s", *bind, *puerto, *modo)
	for {
		c, err := ln.Accept()
		if err != nil {
			log.Fatal(err)
		}
		go atender(c, a)
	}
}
