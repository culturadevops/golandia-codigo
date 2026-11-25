package resumen

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"sync"
)

// El servicio: POST /resumen con el registro en el cuerpo. Un handler por versión.

// HandlerV1 es el servicio de partida.
func HandlerV1(w http.ResponseWriter, r *http.Request) {
	cuerpo, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	salida, err := ResumenV1(cuerpo)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(salida)
}

// HandlerV3 usa ResumenV3, pero sigue leyendo el cuerpo con io.ReadAll: un []byte nuevo
// por petición, que además crece a saltos mientras lee.
func HandlerV3(w http.ResponseWriter, r *http.Request) {
	cuerpo, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	salida, err := ResumenV3(cuerpo)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(salida)
}

// Búferes para reutilizar entre peticiones (lección 5): uno para leer el cuerpo y otro para
// escribir la respuesta. Por separado, porque tienen tamaños muy distintos.
var (
	bufsEntrada = sync.Pool{New: func() any { return new(bytes.Buffer) }}
	bufsSalida  = sync.Pool{New: func() any { return new(bytes.Buffer) }}
)

// HandlerV4 lee el cuerpo en un búfer del pool y escribe la respuesta con otro.
// ⚠ Tiene una condición de carrera: la lección 8 la encuentra con -race.
func HandlerV4(w http.ResponseWriter, r *http.Request) {
	in := bufsEntrada.Get().(*bytes.Buffer)
	in.Reset()
	defer bufsEntrada.Put(in)
	if _, err := in.ReadFrom(r.Body); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	salida, err := ResumenV4(in.Bytes())
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(salida)
}

// ResumenV4 codifica la respuesta en un búfer del pool... y lo devuelve al pool antes de que
// nadie haya usado los bytes. Con una petición a la vez, nunca falla.
func ResumenV4(cuerpo []byte) ([]byte, error) {
	filas, err := agregar(cuerpo)
	if err != nil {
		return nil, err
	}
	out := bufsSalida.Get().(*bytes.Buffer)
	out.Reset()
	defer bufsSalida.Put(out) // ⚠ el búfer vuelve al pool al salir de la función...
	if err := json.NewEncoder(out).Encode(filas); err != nil {
		return nil, err
	}
	return bytes.TrimSuffix(out.Bytes(), []byte("\n")), nil // ...pero estos bytes son SUYOS
}

// HandlerV5 es V4 arreglada: la respuesta se escribe ANTES de devolver el búfer al pool.
func HandlerV5(w http.ResponseWriter, r *http.Request) {
	in := bufsEntrada.Get().(*bytes.Buffer)
	in.Reset()
	defer bufsEntrada.Put(in)
	if _, err := in.ReadFrom(r.Body); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	if err := EscribirV5(w, in.Bytes()); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
	}
}

// EscribirV5 escribe el resumen en w mientras el búfer todavía es suyo.
func EscribirV5(w io.Writer, cuerpo []byte) error {
	filas, err := agregar(cuerpo)
	if err != nil {
		return err
	}
	out := bufsSalida.Get().(*bytes.Buffer)
	out.Reset()
	defer bufsSalida.Put(out)
	if err := json.NewEncoder(out).Encode(filas); err != nil {
		return err
	}
	_, err = w.Write(bytes.TrimSuffix(out.Bytes(), []byte("\n")))
	return err
}
