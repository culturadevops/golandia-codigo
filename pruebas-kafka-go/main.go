// main.go — kafkita: un Kafka de juguete en Go, hablando el protocolo binario de verdad.
//
// Vídeo 224 · «Escribí Kafka en Go y el cliente oficial no notó la diferencia».
//
// ⛔ SIN DEPENDENCIAS. Solo la biblioteca estándar: net, encoding/binary y poco más. El protocolo
// de Kafka es binario, pero es binario SENCILLO: enteros de tamaño fijo en orden de red y cadenas
// con su longitud delante.
//
// 🥇 LA DECISION DE DISEÑO QUE HACE POSIBLE TODO ESTO, y que es el corazón del vídeo:
//
//	En la respuesta a ApiVersions TÚ dices qué versiones soportas. El cliente negocia HACIA ABAJO
//	y usa la más alta que los dos entiendan. Si declaras versiones BAJAS, el cliente deja de usar
//	el formato "flexible" (campos etiquetados y enteros de longitud variable, que llegó en 2020) y
//	pasa a hablar el formato viejo, de longitudes fijas, que se escribe en una tarde.
//
//	No es un truco: es cómo está diseñado el protocolo para que un broker viejo y un cliente nuevo
//	puedan hablar. Aquí se usa a propósito y se cuenta.
//
// ⚠️ LA EXCEPCION QUE HAY QUE SABER: la RESPUESTA a ApiVersions siempre lleva cabecera v0 (solo
// el identificador de correlación), aunque la petición venga en versión moderna. Tiene que ser
// así porque el cliente todavía no sabe qué hablas: es el único mensaje que se contesta antes de
// haber negociado nada.
package main

import (
	"encoding/binary"
	"fmt"
	"io"
	"net"
	"os"
	"sync"
	"time"
)

// Los identificadores de las peticiones que atendemos. Están en la documentación del protocolo.
const (
	apiProduce        = 0
	apiFetch          = 1
	apiListOffsets    = 2
	apiMetadata       = 3
	apiApiVersions    = 18
	apiInitProducerId = 22
)

var nombreApi = map[int16]string{0: "Produce", 1: "Fetch", 2: "ListOffsets",
	3: "Metadata", 18: "ApiVersions", 22: "InitProducerId"}

// ── El almacén. Un Kafka de verdad escribe en disco en segmentos; esto es una lista en memoria.
//
// ⚠️ LA UNIDAD NO ES EL MENSAJE: ES EL LOTE. El productor agrupa varios mensajes en un lote y lo
// manda entero, comprimido y con una suma de comprobación. El broker guarda ese bloque tal cual,
// sin abrirlo. Por eso mandar tres mensajes deja UN lote, no tres registros.
type lote struct {
	base  int64  // el offset del primer mensaje del lote
	n     int32  // cuántos mensajes lleva dentro
	datos []byte // el bloque, tal y como llegó
}

var (
	temas   = map[string][]lote{"pruebas": {}}
	candado sync.Mutex
)

// Posiciones dentro de la cabecera de un lote de registros (formato v2, 61 bytes de cabecera).
const (
	posBaseOffset = 0  // int64
	posNumRegs    = 57 // int32
)

func finDe(t string) int64 {
	l := temas[t]
	if len(l) == 0 {
		return 0
	}
	u := l[len(l)-1]
	return u.base + int64(u.n)
}

// ── ESCRITURA ────────────────────────────────────────────────────────────────
// Un búfer que sabe poner enteros en orden de red y cadenas con la longitud delante.
type buf struct{ b []byte }

func (w *buf) i8(v int8)   { w.b = append(w.b, byte(v)) }
func (w *buf) i16(v int16) { w.b = binary.BigEndian.AppendUint16(w.b, uint16(v)) }
func (w *buf) i32(v int32) { w.b = binary.BigEndian.AppendUint32(w.b, uint32(v)) }
func (w *buf) i64(v int64) { w.b = binary.BigEndian.AppendUint64(w.b, uint64(v)) }

// Una cadena en Kafka es: longitud en 2 bytes, y luego los bytes. Sin terminador nulo.
func (w *buf) str(s string) { w.i16(int16(len(s))); w.b = append(w.b, s...) }
func (w *buf) bytes(b []byte) {
	if b == nil {
		w.i32(-1) // -1 es "nulo". No es lo mismo que vacío, y el cliente lo distingue.
		return
	}
	w.i32(int32(len(b)))
	w.b = append(w.b, b...)
}

// ── LECTURA ──────────────────────────────────────────────────────────────────
type lec struct {
	b []byte
	i int
}

func (r *lec) i16() int16 {
	v := int16(binary.BigEndian.Uint16(r.b[r.i:]))
	r.i += 2
	return v
}

func (r *lec) i32() int32 {
	v := int32(binary.BigEndian.Uint32(r.b[r.i:]))
	r.i += 4
	return v
}

func (r *lec) i8() int8 {
	v := int8(r.b[r.i])
	r.i++
	return v
}

func (r *lec) i64() int64 {
	v := int64(binary.BigEndian.Uint64(r.b[r.i:]))
	r.i += 8
	return v
}

func (r *lec) str() string {
	n := int(r.i16())
	if n < 0 {
		return ""
	}
	s := string(r.b[r.i : r.i+n])
	r.i += n
	return s
}

// ── EL SERVIDOR ──────────────────────────────────────────────────────────────
func main() {
	anuncio := os.Getenv("ANUNCIO") // el host que le decimos al cliente en Metadata
	if anuncio == "" {
		anuncio = "kafkita"
	}
	l, err := net.Listen("tcp", ":9092")
	if err != nil {
		fmt.Println("no puedo escuchar:", err)
		os.Exit(1)
	}
	fmt.Printf("kafkita escuchando en :9092 · me anuncio como %q\n", anuncio)
	for {
		c, err := l.Accept()
		if err != nil {
			continue
		}
		go atender(c, anuncio)
	}
}

func atender(c net.Conn, anuncio string) {
	defer c.Close()
	for {
		// Cada mensaje va precedido de su tamaño en 4 bytes. Sin eso no se sabe dónde acaba:
		// TCP es un chorro de bytes, no una lista de mensajes.
		var cab [4]byte
		if _, err := io.ReadFull(c, cab[:]); err != nil {
			return
		}
		n := int(binary.BigEndian.Uint32(cab[:]))
		cuerpo := make([]byte, n)
		if _, err := io.ReadFull(c, cuerpo); err != nil {
			return
		}

		r := &lec{b: cuerpo}
		api, ver, corr := r.i16(), r.i16(), r.i32()
		cliente := r.str()
		fmt.Printf("  <- %-12s v%-2d corr=%-3d cliente=%q\n",
			nombreApi[api], ver, corr, cliente)

		var resp []byte
		switch api {
		case apiApiVersions:
			resp = respApiVersions(corr)
		case apiMetadata:
			resp = respMetadata(corr, r, anuncio)
		case apiProduce:
			resp = respProduce(corr, r)
		case apiFetch:
			resp = respFetch(corr, r)
		case apiListOffsets:
			resp = respListOffsets(corr, r)
		case apiInitProducerId:
			resp = respInitProducerId(corr)
		default:
			fmt.Printf("     (no la atiendo, cierro)\n")
			return
		}
		if os.Getenv("VOLCADO") == "1" {
			fmt.Printf("     -> %d bytes: % x\n", len(resp), resp)
		}
		var salida buf
		salida.i32(int32(len(resp)))
		salida.b = append(salida.b, resp...)
		if _, err := c.Write(salida.b); err != nil {
			return
		}
	}
}

// ── ApiVersions ──────────────────────────────────────────────────────────────
// 🥇 AQUI SE DECIDE TODO. Al declarar versiones máximas bajas, el cliente negocia hacia abajo y
// habla el formato antiguo, que es el que sabemos escribir.
//
// ⚠️ Y la respuesta lleva cabecera v0: solo el identificador de correlación, nada más. Es el
// único mensaje del protocolo que se contesta sin haber negociado nada todavía.
func respApiVersions(corr int32) []byte {
	var w buf
	w.i32(corr) // cabecera v0
	w.i16(0)    // código de error: 0 = todo bien
	w.i32(6)    // cuántas entradas vienen

	// api, versión mínima, versión MÁXIMA que decimos soportar
	for _, e := range [][3]int16{
		// ⚠️ Produce llega hasta la 3 y NO menos, y esto es un hallazgo del vídeo: el productor
		// idempotente (el de serie desde Kafka 3.0) EXIGE la versión 3 como mínimo, porque los
		// campos del identificador de productor y el número de secuencia no existen antes.
		// O sea: la negociación hacia abajo tiene un suelo, y lo pone la funcionalidad que pidas.
		{apiProduce, 0, 3},
		{apiFetch, 0, 11},
		{apiListOffsets, 0, 1},
		{apiMetadata, 0, 3},
		{apiApiVersions, 0, 2},
		{apiInitProducerId, 0, 0},
	} {
		w.i16(e[0])
		w.i16(e[1])
		w.i16(e[2])
	}
	w.i32(0) // throttle_time_ms
	return w.b
}

// ── Metadata ─────────────────────────────────────────────────────────────────
// ⚠️ LA TRAMPA MAS FAMOSA DE KAFKA EN LA VIDA REAL vive aquí. En esta respuesta le decimos al
// cliente a qué HOST y PUERTO tiene que conectarse para hablar con cada partición. Si aquí
// anunciamos un nombre que el cliente no puede resolver, el cliente se conecta bien a nosotros,
// recibe esta respuesta, y a partir de ahí intenta ir al nombre anunciado y se pierde.
// Es el motivo de casi todos los "no me conecto a Kafka" con Docker.
func respMetadata(corr int32, r *lec, anuncio string) []byte {
	pedidos := []string{}
	if r.i < len(r.b) {
		n := r.i32()
		for i := int32(0); i < n && r.i < len(r.b); i++ {
			pedidos = append(pedidos, r.str())
		}
	}
	candado.Lock()
	if len(pedidos) == 0 {
		for t := range temas {
			pedidos = append(pedidos, t)
		}
	}
	for _, t := range pedidos {
		if _, hay := temas[t]; !hay {
			temas[t] = []lote{} // creación automática, como Kafka con auto.create
		}
	}
	candado.Unlock()

	var w buf
	w.i32(corr)
	w.i32(0) // throttle
	w.i32(1) // un solo broker: nosotros
	w.i32(0) // node id
	w.str(anuncio)
	w.i32(9092)
	w.str("") // rack
	w.str("kafkita-cluster")
	w.i32(0) // controller id
	w.i32(int32(len(pedidos)))
	for _, t := range pedidos {
		w.i16(0) // error del tema
		w.str(t)
		w.i8(0)  // ¿interno? no
		w.i32(1) // una partición
		w.i16(0) // error de la partición
		w.i32(0) // partición 0
		w.i32(0) // líder: nosotros
		w.i32(1)
		w.i32(0) // réplicas: [0]
		w.i32(1)
		w.i32(0) // en sincronía: [0]
	}
	return w.b
}

// ── Produce ──────────────────────────────────────────────────────────────────
// El cuerpo del lote de registros es un formato propio con su versión, sus compresiones y su
// suma de comprobación. Aquí se guarda tal cual llega y se cuenta cuántos bytes son: lo que
// demuestra el vídeo es el PROTOCOLO, no el formato de almacenamiento.
func respProduce(corr int32, r *lec) []byte {
	// Desde la versión 3 la petición empieza con el identificador de transacción, que puede ser
	// nulo (longitud -1). El productor normal manda nulo; solo lo llena quien usa transacciones.
	r.str() // transactional_id, nulo casi siempre
	r.i16() // acks
	r.i32() // timeout
	nt := r.i32()
	type res struct {
		tema  string
		base  int64
		bytes int
	}
	var out []res
	for i := int32(0); i < nt; i++ {
		tema := r.str()
		np := r.i32()
		for j := int32(0); j < np; j++ {
			r.i32() // partición
			n := int(r.i32())
			datos := r.b[r.i : r.i+n]
			r.i += n
			// Cuántos mensajes trae el lote, leído de su cabecera. El productor los agrupa: tres
			// mensajes seguidos llegan en UN lote, no en tres peticiones.
			var cuantos int32 = 1
			if len(datos) >= 61 {
				cuantos = int32(binary.BigEndian.Uint32(datos[posNumRegs:]))
			}
			candado.Lock()
			base := finDe(tema)
			copia := append([]byte(nil), datos...)
			// 🥇 EL BROKER REESCRIBE EL LOTE. El productor manda el lote con offset base 0, porque
			// no sabe qué offsets le van a tocar: eso lo decide el que escribe en el registro. Así
			// que aquí se pisan esos 8 bytes con el offset de verdad. Si no lo haces, el consumidor
			// pide desde una posición y recibe un lote que dice empezar en otra, y se lía.
			if len(copia) >= 8 {
				binary.BigEndian.PutUint64(copia[posBaseOffset:], uint64(base))
			}
			temas[tema] = append(temas[tema], lote{base: base, n: cuantos, datos: copia})
			candado.Unlock()
			out = append(out, res{tema, base, n})
			fmt.Printf("     guardado en %q: %d bytes · %d mensajes · offsets %d..%d\n",
				tema, n, cuantos, base, base+int64(cuantos)-1)
		}
	}
	var w buf
	w.i32(corr)
	w.i32(int32(len(out)))
	for _, o := range out {
		w.str(o.tema)
		w.i32(1)
		w.i32(0)      // partición
		w.i16(0)      // sin error
		w.i64(o.base) // el offset que le asignamos
		w.i64(-1)     // log_append_time
	}
	w.i32(0) // throttle
	return w.b
}

// ── Fetch ────────────────────────────────────────────────────────────────────
// Se atiende en la VERSION 11, y el motivo es el hallazgo más útil de todo el vídeo.
//
// ⛔ Con Fetch declarado hasta la 3, el cliente oficial de Kafka 4 no da ningún error: escribe en
// su registro «Fetch(1): 0 to 3 [unusable: node too old]» y NO manda la petición nunca. El
// consumidor se queda esperando y muere por tiempo agotado sin decir por qué.
//
// ⚠️ Y compara los dos fallos, porque son la misma causa con dos caras:
//
//	Produce demasiado viejo  ->  excepción ruidosa: «The node does not support PRODUCE...»
//	Fetch demasiado viejo    ->  silencio absoluto
//
// Cuando un consumidor de Kafka «no lee y no dice nada», mira las versiones antes que la red.
//
// La 11 es la última antes de que Fetch pase al formato flexible en la 12, así que es la más
// nueva que se puede escribir sin implementar enteros de longitud variable.
func respFetch(corr int32, r *lec) []byte {
	r.i32()            // replica_id
	maxWait := r.i32() // max_wait_ms: cuánto está dispuesto a esperar el cliente
	r.i32()            // min_bytes
	r.i32()            // max_bytes          (desde la v3)
	r.i8()             // isolation_level    (desde la v4)
	r.i32()            // session_id         (desde la v7)
	r.i32()            // session_epoch      (desde la v7)
	nt := r.i32()
	type res struct {
		tema  string
		datos []byte
		alto  int64
	}
	var out []res
	for i := int32(0); i < nt; i++ {
		tema := r.str()
		np := r.i32()
		for j := int32(0); j < np; j++ {
			r.i32()          // partición
			r.i32()          // current_leader_epoch (desde la v9)
			desde := r.i64() // el offset desde el que quiere leer
			r.i64()          // log_start_offset     (desde la v5)
			r.i32()          // partition_max_bytes
			// ⏳ ESPERA LARGA. Un broker de verdad NO contesta enseguida cuando no hay nada: se
			// queda con la petición abierta hasta que llegue algo o se acabe el tiempo que pidió
			// el cliente. Sin esto, el consumidor pregunta en bucle cerrado: en la primera
			// corrida llegó a 316.000 peticiones en diez segundos. Con esto, dos por segundo.
			espera := time.Duration(maxWait) * time.Millisecond
			if espera > 5*time.Second {
				espera = 5 * time.Second
			}
			limite := time.Now().Add(espera)
			var junto []byte
			var alto int64
			for {
				candado.Lock()
				lotes := temas[tema]
				junto = nil
				for _, l := range lotes {
					if l.base+int64(l.n) > desde {
						junto = append(junto, l.datos...)
					}
				}
				alto = finDe(tema)
				candado.Unlock()
				if len(junto) > 0 || time.Now().After(limite) {
					break
				}
				time.Sleep(50 * time.Millisecond)
			}
			out = append(out, res{tema, junto, alto})
			fmt.Printf("     %q pide desde %d · le mando %d bytes (tope %d)\n",
				tema, desde, len(junto), alto)
		}
	}
	var w buf
	w.i32(corr)
	w.i32(0) // throttle_time_ms
	w.i16(0) // error_code de toda la petición (desde la v7)
	w.i32(0) // session_id                     (desde la v7)
	w.i32(int32(len(out)))
	for _, o := range out {
		w.str(o.tema)
		w.i32(1)
		w.i32(0)      // partición
		w.i16(0)      // sin error
		w.i64(o.alto) // high_watermark
		w.i64(o.alto) // last_stable_offset     (desde la v4)
		w.i64(0)      // log_start_offset       (desde la v5)
		w.i32(0)      // sin transacciones abortadas
		w.i32(-1)     // preferred_read_replica (desde la v11)
		w.bytes(o.datos)
	}
	return w.b
}

// ── ListOffsets ──────────────────────────────────────────────────────────────
func respListOffsets(corr int32, r *lec) []byte {
	r.i32() // replica_id
	nt := r.i32()
	type res struct {
		tema string
		off  int64
	}
	var out []res
	for i := int32(0); i < nt; i++ {
		tema := r.str()
		np := r.i32()
		for j := int32(0); j < np; j++ {
			r.i32() // partición
			marca := int64(binary.BigEndian.Uint64(r.b[r.i:]))
			r.i += 8
			candado.Lock()
			off := int64(0)
			if marca == -1 { // -1 = "el final"; -2 = "el principio"
				off = finDe(tema)
			}
			candado.Unlock()
			out = append(out, res{tema, off})
			fmt.Printf("     ListOffsets %q marca=%d -> offset %d\n", tema, marca, off)
		}
	}
	var w buf
	w.i32(corr)
	w.i32(int32(len(out)))
	for _, o := range out {
		w.str(o.tema)
		w.i32(1)
		w.i32(0)
		w.i16(0)
		w.i64(-1) // timestamp
		w.i64(o.off)
	}
	return w.b
}

// ── InitProducerId ───────────────────────────────────────────────────────────
// ⚠️ HACE FALTA AUNQUE NO LO PAREZCA. Desde Kafka 3.0 el productor oficial es IDEMPOTENTE por
// defecto: antes de mandar un solo mensaje pide un identificador de productor, y con él numera
// sus lotes para que el broker pueda descartar duplicados si hay un reintento. Si no contestas a
// esto, el productor aborta con "The node does not support INIT_PRODUCER_ID" y no manda nada.
// Aquí devolvemos un identificador fijo: no deduplicamos nada, y se dice en el vídeo.
func respInitProducerId(corr int32) []byte {
	var w buf
	w.i32(corr)
	w.i32(0)    // throttle
	w.i16(0)    // sin error
	w.i64(1000) // producer_id, inventado y constante
	w.i16(0)    // producer_epoch
	return w.b
}
