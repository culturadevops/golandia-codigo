package main

// De donde salen de verdad los registros de tu funcion.
//
// El episodio 7 dejo un hueco escrito con todas las letras: «lo que imprima la funcion se queda
// en su contenedor». Este fichero es el que va a buscarlo.
//
// 🥇 Y aqui esta el descubrimiento tecnico, que no se ve en ninguna documentacion de AWS porque
// no es de AWS: la salida de un contenedor NO viene en texto plano. Viene en MARCOS. Docker
// mete stdout y stderr por la MISMA conexion, y para que se puedan separar al otro lado le pone
// a cada trozo una cabecera de ocho bytes:
//
//	byte 0     de que flujo viene: 0 stdin, 1 stdout, 2 stderr
//	bytes 1-3  cero, sin usar
//	bytes 4-7  cuantos bytes vienen detras, en big endian
//
// Si te lo lees como texto, te encuentras basura binaria cada pocas lineas y piensas que el log
// esta corrupto. No lo esta: estas leyendo el sobre.
//
// ⚠️ Y hay un caso en el que NO hay marcos: si el contenedor se creo con `Tty: true`. Con
// terminal solo hay un flujo, asi que no hay nada que separar y Docker manda el texto tal cual.
// El mismo endpoint devuelve dos formatos distintos segun como se creo el contenedor, y no te
// avisa. Aqui se miran los primeros ocho bytes para saber cual de los dos es.
//
// ⚠️ Un marco NO es una linea. Un marco puede traer media linea, tres lineas, o una linea
// partida justo por la mitad de una palabra. Hay que ir juntando por flujo hasta ver un salto.

import (
	"bufio"
	"bytes"
	"context"
	"encoding/binary"
	"errors"
	"io"
	"net/http"
	"strings"
	"time"
)

// topeMarco es una defensa, no un limite de Docker: si la cabecera viene mal, el numero de
// bytes que dice puede ser cualquier cosa y reservarlo seria pedirle a la maquina un gigabyte
// por un byte torcido.
const topeMarco = 16 << 20

// ConRegistros engancha el lanzador a CloudWatch Logs. Sin esto el lanzador sigue funcionando
// igual y los registros se quedan donde estaban: dentro del contenedor.
func (l *Lanzador) ConRegistros(r *Registros) *Lanzador {
	if l == nil {
		return nil
	}
	l.reg = r
	return l
}

// ConVida le da al lanzador el contexto que marca cuanto vive el proceso.
//
// ⚠️ Hace falta porque seguir los registros de un contenedor es una conexion ABIERTA, que dura
// lo que dure el contenedor. Si usara el contexto de la peticion que provoco el arranque, se
// cortaria en cuanto esa invocacion contestara — o sea, casi siempre antes de que la funcion
// llegue a imprimir nada.
func (l *Lanzador) ConVida(ctx context.Context) *Lanzador {
	if l == nil {
		return nil
	}
	l.vida = ctx
	return l
}

func (l *Lanzador) contexto() context.Context {
	if l.vida != nil {
		return l.vida
	}
	return context.Background()
}

// seguir abre la salida del contenedor y va metiendo cada linea en su flujo de registro.
//
// `follow=1` deja la conexion abierta hasta que el contenedor muere: no es un sondeo, es el
// demonio empujando. `timestamps=1` hace que cada linea venga con la hora que le puso DOCKER,
// que es la buena: la hora en que se imprimio, no la hora en que nos llego.
func (l *Lanzador) seguir(ctx context.Context, id, grupo, flujo string) {
	ruta := "/containers/" + id + "/logs?stdout=1&stderr=1&follow=1&timestamps=1&tail=all"
	pet, err := http.NewRequestWithContext(ctx, http.MethodGet, l.base+"/"+versionAPI+ruta, nil)
	if err != nil {
		l.registro("  -> no puedo pedir los registros de %s: %v", corto(id), err)
		return
	}
	res, err := l.cli.Do(pet)
	if err != nil {
		// Que se corte no es un fallo: el contenedor se apago, o el proceso esta cerrando.
		if ctx.Err() == nil {
			l.registro("  -> se corto la salida de %s: %v", corto(id), err)
		}
		return
	}
	defer res.Body.Close()
	if res.StatusCode >= 400 {
		crudo, _ := io.ReadAll(res.Body)
		l.registro("  -> docker no me da los registros de %s: %d %s", corto(id),
			res.StatusCode, strings.TrimSpace(string(crudo)))
		return
	}

	err = demultiplexar(res.Body, func(_ byte, linea []byte) {
		marca, mensaje := partirMarca(string(linea))
		l.reg.Anexar(grupo, flujo, marca, mensaje)
	})
	if err != nil && ctx.Err() == nil && !errors.Is(err, io.EOF) {
		l.registro("  -> los registros de %s se cortaron raro: %v", corto(id), err)
	}
}

// demultiplexar separa los marcos de Docker y entrega LINEAS enteras.
//
// El `flujo` que llega al callback es 1 para stdout y 2 para stderr. Aqui no se usa para nada
// —CloudWatch tampoco distingue— pero se pasa porque tirarlo seria perder informacion que ya
// estaba, y con eso se pueden pintar los errores en rojo el dia que apetezca.
func demultiplexar(r io.Reader, en func(flujo byte, linea []byte)) error {
	br := bufio.NewReader(r)
	cab, err := br.Peek(8)
	if err != nil || !esMarco(cab) {
		// Sin marcos: o el contenedor tiene terminal, o no hay ni ocho bytes que mirar. En los
		// dos casos lo que hay es texto.
		if err != nil && !errors.Is(err, io.EOF) && !errors.Is(err, bufio.ErrBufferFull) {
			return err
		}
		return porLineas(br, func(l []byte) { en(1, l) })
	}

	restos := map[byte][]byte{}
	c := make([]byte, 8)
	for {
		if _, err := io.ReadFull(br, c); err != nil {
			if errors.Is(err, io.EOF) || errors.Is(err, io.ErrUnexpectedEOF) {
				break
			}
			return err
		}
		cual := c[0]
		n := binary.BigEndian.Uint32(c[4:])
		if n > topeMarco {
			return errors.New("marco de docker imposible: dice traer demasiados bytes")
		}
		datos := make([]byte, n)
		if _, err := io.ReadFull(br, datos); err != nil {
			if !errors.Is(err, io.EOF) && !errors.Is(err, io.ErrUnexpectedEOF) {
				return err
			}
			// Lo que haya llegado vale: puede ser la ultima linea sin su salto.
			datos = datos[:0]
		}
		buf := append(restos[cual], datos...)
		for {
			i := bytes.IndexByte(buf, '\n')
			if i < 0 {
				break
			}
			en(cual, copiar(recortar(buf[:i])))
			buf = buf[i+1:]
		}
		restos[cual] = copiar(buf)
	}
	// Lo que quedo a medias tambien es una linea: un proceso puede morirse sin salto final.
	// ⚠️ En orden fijo (1 y luego 2) porque recorrer un mapa en Go es recorrerlo al azar, y una
	// salida que cambia de orden entre corridas es una prueba que falla un martes.
	for _, cual := range []byte{1, 2} {
		if len(restos[cual]) > 0 {
			en(cual, copiar(recortar(restos[cual])))
		}
	}
	return nil
}

// esMarco dice si esos ocho bytes tienen pinta de cabecera de Docker.
//
// ⚠️ No hay forma de saberlo con certeza, y conviene decirlo en voz alta: una linea de texto que
// empezara por los bytes correctos engañaria a esto. En la practica no pasa —el byte 0 tiene que
// ser 0, 1 o 2, y los tres siguientes cero, o sea cuatro caracteres de control seguidos— pero
// «en la practica no pasa» no es «no puede pasar».
func esMarco(cab []byte) bool {
	return len(cab) >= 8 && cab[0] <= 2 && cab[1] == 0 && cab[2] == 0 && cab[3] == 0
}

func porLineas(r io.Reader, en func([]byte)) error {
	sc := bufio.NewScanner(r)
	sc.Buffer(make([]byte, 0, 64*1024), topeMarco)
	for sc.Scan() {
		en(copiar(recortar(sc.Bytes())))
	}
	return sc.Err()
}

// recortar quita el retorno de carro que dejan los programas que escriben con final de linea de
// Windows. Sin esto, cada mensaje del registro acaba en un caracter invisible.
func recortar(b []byte) []byte {
	if len(b) > 0 && b[len(b)-1] == '\r' {
		return b[:len(b)-1]
	}
	return b
}

func copiar(b []byte) []byte { return append([]byte(nil), b...) }

// partirMarca separa la hora que puso Docker del mensaje.
//
// Con `timestamps=1` cada linea llega asi:
//
//	2026-08-25T15:04:05.123456789Z lo que imprimio tu funcion
//
// ⚠️ Y esa hora es la del CONTENEDOR, no la nuestra. Es la correcta: si la pusieramos nosotros
// al recibirla, un contenedor que escribe rapido y una red lenta te ordenarian mal el registro.
// El precio es que si el contenedor tiene la hora mal, sus apuntes pueden caer fuera de la
// ventana y desaparecer sin decir nada. Vease `Poner` en registros.go.
func partirMarca(linea string) (int64, string) {
	if i := strings.IndexByte(linea, ' '); i > 0 {
		if t, err := time.Parse(time.RFC3339Nano, linea[:i]); err == nil {
			return t.UnixMilli(), linea[i+1:]
		}
	}
	return time.Now().UnixMilli(), linea
}
