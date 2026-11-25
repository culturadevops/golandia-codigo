package main

// El lanzador. Episodio 7, y es lo que separa «emular Lambda» de «ser Lambda».
//
// En el episodio 5 la funcion se ejecutaba de verdad, con la biblioteca oficial de Amazon. Pero
// el contenedor lo levantabas TU, a mano, con un `docker run`. Eso esta bien para entender el
// protocolo y es una mentira para usarlo: en Lambda de verdad tu no levantas nada. Llamas, y
// alguien levanta lo que haga falta.
//
// Ese «alguien» es este fichero.
//
// 🥇 Y el descubrimiento que lo hace posible: la API de Docker es HTTP. Nada mas. Un socket de
// dominio Unix en `/var/run/docker.sock` con peticiones HTTP normales encima. No hace falta el
// SDK de Docker ni ninguna dependencia: la biblioteca estandar de Go ya sabe hablar por un
// socket, y `net/http` no pregunta por donde va la conexion.
//
//	POST /containers/create   -> crear el contenedor
//	POST /containers/{id}/start
//	DELETE /containers/{id}?force=1
//
// ⛔ AVISO SERIO, y va en el video: montar el socket de Docker dentro de un contenedor le da a
// ese contenedor control del demonio ENTERO de la maquina. Puede crear, borrar y montar lo que
// quiera. Es lo que hace LocalStack, es lo que hace esto, y es exactamente por eso que un doble
// de pruebas no se pone en un servidor compartido.

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"sort"
	"strings"
	"sync"
	"time"
)

// versionAPI la fija el cliente, no el servidor. Sin ella el demonio contesta con la suya, que
// puede ser mas nueva que lo que este codigo entiende.
const versionAPI = "v1.41"

// etiqueta marca los contenedores que hemos levantado nosotros. Sin ella no podriamos
// distinguir «un contenedor de esta funcion» de «un contenedor que puso alguien a mano», y
// borrar por nombre es como borrar por adivinanza.
const etiqueta = "nubecita.funcion"

type Lanzador struct {
	cli      *http.Client
	base     string // "http://docker": el host da igual, la conexion va por el socket
	red      string // la red de Docker donde vive todo
	yo       string // como me ven los contenedores: "nubecita:4566"
	max      int    // tope de contenedores por funcion
	ocioso   time.Duration
	registro func(formato string, args ...any)

	// A donde van a parar los registros de lo que levantamos, y cuanto vive el proceso. Los
	// dos se ponen con `ConRegistros` y `ConVida`; sin ellos el lanzador hace lo de siempre.
	// Vease bitacora.go.
	reg  *Registros
	vida context.Context

	mu         sync.Mutex
	porFuncion map[string][]string // funcion -> ids de contenedor
	sec        int
	arrancados int
	apagados   int
}

// ClienteDocker devuelve un cliente HTTP que habla por un socket en vez de por la red.
//
// ⚠️ Fijate en lo poco que hace falta: se cambia como se ABRE la conexion y se deja igual todo
// lo demas. `net/http` no sabe ni le importa que al otro lado no haya una direccion IP.
func ClienteDocker(socket string) *http.Client {
	return &http.Client{
		Transport: &http.Transport{
			DialContext: func(ctx context.Context, _, _ string) (net.Conn, error) {
				var d net.Dialer
				return d.DialContext(ctx, "unix", socket)
			},
		},
		// Sin tiempo limite global: crear un contenedor puede tardar si la imagen es grande.
		// Cada llamada trae su propio contexto.
		Timeout: 0,
	}
}

func NuevoLanzador(cli *http.Client, base, red, yo string, max int, ocioso time.Duration,
	registro func(string, ...any)) *Lanzador {
	if registro == nil {
		registro = func(string, ...any) {}
	}
	if max <= 0 {
		max = 4
	}
	return &Lanzador{cli: cli, base: strings.TrimSuffix(base, "/"), red: red, yo: yo,
		max: max, ocioso: ocioso, registro: registro,
		porFuncion: map[string][]string{}}
}

// ── la decision ──────────────────────────────────────────────────────────

// Asegurar mira si hace falta levantar un runtime para esta funcion, y lo levanta.
//
// 🥇 La regla es de una linea: si no hay NINGUN runtime esperando trabajo, hace falta otro.
// Y de esa linea sale, sin proponerselo, el escalado de Lambda:
//
//	una invocacion con la funcion parada   -> se levanta uno, y esa llamada paga el frio
//	una segunda mientras la primera trabaja -> tampoco hay libres: se levanta otro
//	una tercera cuando ya hay uno libre     -> no se levanta nada
//
// O sea que una rafaga de diez llamadas te levanta contenedores hasta el tope. Eso no es un
// efecto secundario: es lo que significa «concurrencia» en la factura de Lambda.
func (l *Lanzador) Asegurar(ctx context.Context, fn *Funcion) {
	if l == nil || fn.Imagen == "" {
		return
	}
	l.mu.Lock()
	defer l.mu.Unlock()

	if fn.Libres() > 0 {
		return // ya hay quien atienda: no se toca nada
	}
	if len(l.porFuncion[fn.Nombre]) >= l.max {
		l.registro("  -> %s ya tiene %d contenedores (el tope): esta invocacion espera turno",
			fn.Nombre, l.max)
		return
	}
	if err := l.arrancar(ctx, fn); err != nil {
		// Que no se pueda levantar no es motivo para tirar la invocacion: puede que haya un
		// runtime tuyo, levantado a mano, a punto de pedir trabajo. Se registra y se sigue.
		l.registro("  -> no puedo levantar un runtime para %s: %v", fn.Nombre, err)
	}
}

func (l *Lanzador) arrancar(ctx context.Context, fn *Funcion) error {
	l.sec++
	nombre := fmt.Sprintf("nubecita-%s-%d", fn.Nombre, l.sec)

	// 🥇 La primera de estas lineas es todo el episodio 5 metido en una variable de entorno. La
	// biblioteca oficial de Amazon la lee al arrancar y viene a pedirnos trabajo. Las demas son
	// las que configuraste tu al crear la funcion: en Lambda el codigo no cambia entre
	// entornos, cambian estas.
	//
	// ⚠️ Van ORDENADAS. Un mapa en Go se recorre al azar, y sin ordenar dos contenedores de la
	// misma funcion se crearian con la misma configuracion escrita en distinto orden — feo de
	// leer y desagradable de comparar.
	env := []string{"AWS_LAMBDA_RUNTIME_API=" + l.yo + "/f/" + fn.Nombre}
	claves := make([]string, 0, len(fn.Entorno))
	for k := range fn.Entorno {
		claves = append(claves, k)
	}
	sort.Strings(claves)
	for _, k := range claves {
		env = append(env, k+"="+fn.Entorno[k])
	}

	cuerpo := map[string]any{
		"Image":  fn.Imagen,
		"Env":    env,
		"Labels": map[string]string{etiqueta: fn.Nombre},
		"HostConfig": map[string]any{
			"NetworkMode": l.red,
			// ⚠️ AutoRemove NO: si el contenedor se borra solo al morir, un fallo de arranque
			// desaparece antes de que puedas leer por que. Los borramos nosotros, cuando toca.
			"AutoRemove": false,
		},
	}

	var creado struct {
		ID       string   `json:"Id"`
		Warnings []string `json:"Warnings"`
	}
	if err := l.llamar(ctx, http.MethodPost, "/containers/create?name="+url.QueryEscape(nombre),
		cuerpo, &creado); err != nil {
		return err
	}
	for _, w := range creado.Warnings {
		l.registro("  -> docker avisa: %s", w)
	}
	if err := l.llamar(ctx, http.MethodPost, "/containers/"+creado.ID+"/start", nil, nil); err != nil {
		_ = l.borrar(ctx, creado.ID)
		return err
	}

	// Y en cuanto esta vivo, alguien se pone a escuchar lo que imprima. El flujo se crea aqui,
	// no cuando llega la primera linea: asi `DescribeLogStreams` ve el contenedor aunque no
	// haya dicho nada todavia, que es justo cuando mas falta hace mirarlo.
	if l.reg != nil {
		grupo := grupoDeFuncion(fn.Nombre)
		flujo := nombreDeFlujo(time.Now(), creado.ID)
		l.reg.Asegurar(grupo, flujo)
		go l.seguir(l.contexto(), creado.ID, grupo, flujo)
	}

	l.porFuncion[fn.Nombre] = append(l.porFuncion[fn.Nombre], creado.ID)
	l.arrancados++
	l.registro("  -> levantado %s (%s) para %s · van %d", nombre, corto(creado.ID), fn.Nombre,
		len(l.porFuncion[fn.Nombre]))
	return nil
}

// ── recoger lo que sobra ─────────────────────────────────────────────────

// Recoger apaga los contenedores de las funciones que llevan un rato sin usarse.
//
// 🥇 Y esto es lo que hace que el arranque en frio VUELVA. Una funcion que nadie llama en un
// rato se queda sin contenedores, y la siguiente llamada vuelve a pagarlo. En Amazon no te lo
// dicen con estas palabras, pero es exactamente lo que pasa: nadie te guarda una maquina
// encendida por si acaso.
func (l *Lanzador) Recoger(ctx context.Context, fns []*Funcion) {
	if l == nil || l.ocioso <= 0 {
		return
	}
	ahora := time.Now().UTC()
	for _, fn := range fns {
		ultima := fn.Ultima()
		if ultima.IsZero() || ahora.Sub(ultima) < l.ocioso {
			continue
		}
		l.mu.Lock()
		ids := l.porFuncion[fn.Nombre]
		delete(l.porFuncion, fn.Nombre)
		l.mu.Unlock()
		for _, id := range ids {
			if err := l.borrar(ctx, id); err != nil {
				l.registro("  -> no puedo apagar %s: %v", corto(id), err)
				continue
			}
			l.mu.Lock()
			l.apagados++
			l.mu.Unlock()
			l.registro("  -> apagado %s: %s llevaba %s sin invocarse", corto(id), fn.Nombre,
				ahora.Sub(ultima).Round(time.Second))
		}
	}
}

// Vigilar corre `Recoger` cada tanto hasta que el contexto muera.
func (l *Lanzador) Vigilar(ctx context.Context, cada time.Duration, listar func() []*Funcion) {
	if l == nil || l.ocioso <= 0 {
		return
	}
	go func() {
		t := time.NewTicker(cada)
		defer t.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-t.C:
				l.Recoger(ctx, listar())
			}
		}
	}()
}

// Apagar quita todos los contenedores que hemos levantado. Se llama al cerrar y al borrar una
// funcion: si no, quedan contenedores huerfanos de una nube que ya no existe.
func (l *Lanzador) Apagar(ctx context.Context, funcion string) {
	if l == nil {
		return
	}
	l.mu.Lock()
	var ids []string
	if funcion == "" {
		for k, v := range l.porFuncion {
			ids = append(ids, v...)
			delete(l.porFuncion, k)
		}
	} else {
		ids = l.porFuncion[funcion]
		delete(l.porFuncion, funcion)
	}
	l.mu.Unlock()
	sort.Strings(ids)
	for _, id := range ids {
		if err := l.borrar(ctx, id); err != nil {
			l.registro("  -> no puedo apagar %s: %v", corto(id), err)
			continue
		}
		l.mu.Lock()
		l.apagados++
		l.mu.Unlock()
		l.registro("  -> apagado %s", corto(id))
	}
}

// Cuenta dice cuantos contenedores hay levantados de una funcion. Lo miran las pruebas y el
// registro; no decide nada.
func (l *Lanzador) Cuenta(funcion string) int {
	if l == nil {
		return 0
	}
	l.mu.Lock()
	defer l.mu.Unlock()
	return len(l.porFuncion[funcion])
}

func (l *Lanzador) Cifras() (arrancados, apagados int) {
	if l == nil {
		return 0, 0
	}
	l.mu.Lock()
	defer l.mu.Unlock()
	return l.arrancados, l.apagados
}

// ── hablar con el demonio ────────────────────────────────────────────────

func (l *Lanzador) borrar(ctx context.Context, id string) error {
	return l.llamar(ctx, http.MethodDelete, "/containers/"+id+"?force=1&v=1", nil, nil)
}

// llamar hace una peticion a la API de Docker. Es HTTP normal: por eso cabe aqui.
func (l *Lanzador) llamar(ctx context.Context, metodo, ruta string, envio, recibo any) error {
	var cuerpo io.Reader
	if envio != nil {
		crudo, err := json.Marshal(envio)
		if err != nil {
			return err
		}
		cuerpo = bytes.NewReader(crudo)
	}
	pet, err := http.NewRequestWithContext(ctx, metodo, l.base+"/"+versionAPI+ruta, cuerpo)
	if err != nil {
		return err
	}
	if envio != nil {
		pet.Header.Set("Content-Type", "application/json")
	}
	res, err := l.cli.Do(pet)
	if err != nil {
		return err
	}
	defer res.Body.Close()
	crudo, _ := io.ReadAll(res.Body)
	if res.StatusCode >= 400 {
		// Docker devuelve {"message":"..."} y ese mensaje es util de verdad: dice si falta la
		// imagen, si el nombre esta cogido o si la red no existe.
		var e struct {
			Message string `json:"message"`
		}
		_ = json.Unmarshal(crudo, &e)
		if e.Message == "" {
			e.Message = strings.TrimSpace(string(crudo))
		}
		return fmt.Errorf("docker %s %s: %d %s", metodo, ruta, res.StatusCode, e.Message)
	}
	if recibo != nil && len(crudo) > 0 {
		return json.Unmarshal(crudo, recibo)
	}
	return nil
}

func corto(id string) string {
	if len(id) > 12 {
		return id[:12]
	}
	return id
}
