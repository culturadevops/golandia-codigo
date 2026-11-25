// notas-api: la API de juguete contra la que trabaja el provider (golandia 675).
// Notas en memoria, JSON por HTTP. Solo la biblioteca estándar.
//
//	POST   /notas        crea  (201)
//	GET    /notas/{id}   lee   (200 / 404)
//	PUT    /notas/{id}   cambia (200 / 404)
//	DELETE /notas/{id}   borra (204 / 404)
//
// Dos cosas que hace a propósito, como muchas API reales:
//   - guarda las etiquetas ORDENADAS, no en el orden en que llegaron;
//   - la clave se guarda, pero nunca se devuelve.
package main

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
	"sort"
	"strconv"
	"sync"
	"time"
)

type Nota struct {
	ID        string   `json:"id"`
	Titulo    string   `json:"titulo"`
	Texto     string   `json:"texto"`
	Etiquetas []string `json:"etiquetas"`
	Clave     string   `json:"clave,omitempty"` // solo de entrada
	Creada    string   `json:"creada"`
	Version   int64    `json:"version"`
}

var (
	mu    sync.Mutex
	notas = map[string]*Nota{}
	sig   = 1
)

// La respuesta nunca lleva la clave.
func publica(n *Nota) Nota {
	c := *n
	c.Clave = ""
	return c
}

func responder(w http.ResponseWriter, code int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	if v != nil {
		json.NewEncoder(w).Encode(v)
	}
}

func crear(w http.ResponseWriter, r *http.Request) {
	var n Nota
	if err := json.NewDecoder(r.Body).Decode(&n); err != nil || n.Titulo == "" {
		responder(w, 400, map[string]string{"error": "falta titulo"})
		return
	}
	mu.Lock()
	defer mu.Unlock()
	n.ID = strconv.Itoa(sig)
	sig++
	sort.Strings(n.Etiquetas)
	n.Creada = time.Now().UTC().Format(time.RFC3339)
	n.Version = 1
	notas[n.ID] = &n
	responder(w, 201, publica(&n))
}

func leer(w http.ResponseWriter, r *http.Request) {
	mu.Lock()
	defer mu.Unlock()
	n, ok := notas[r.PathValue("id")]
	if !ok {
		responder(w, 404, map[string]string{"error": "no existe"})
		return
	}
	responder(w, 200, publica(n))
}

func cambiar(w http.ResponseWriter, r *http.Request) {
	var in Nota
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil || in.Titulo == "" {
		responder(w, 400, map[string]string{"error": "falta titulo"})
		return
	}
	mu.Lock()
	defer mu.Unlock()
	n, ok := notas[r.PathValue("id")]
	if !ok {
		responder(w, 404, map[string]string{"error": "no existe"})
		return
	}
	n.Titulo, n.Texto, n.Etiquetas = in.Titulo, in.Texto, in.Etiquetas
	sort.Strings(n.Etiquetas)
	if in.Clave != "" {
		n.Clave = in.Clave
	}
	n.Version++
	responder(w, 200, publica(n))
}

func borrar(w http.ResponseWriter, r *http.Request) {
	mu.Lock()
	defer mu.Unlock()
	id := r.PathValue("id")
	if _, ok := notas[id]; !ok {
		responder(w, 404, map[string]string{"error": "no existe"})
		return
	}
	delete(notas, id)
	responder(w, 204, nil)
}

// Cada petición queda en el log: así se ve qué llamadas hace Terraform en cada comando.
type grabadora struct {
	http.ResponseWriter
	code int
}

func (g *grabadora) WriteHeader(c int) { g.code = c; g.ResponseWriter.WriteHeader(c) }

func registrar(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		g := &grabadora{w, 200}
		h.ServeHTTP(g, r)
		log.Printf("%s %s %d", r.Method, r.URL.Path, g.code)
	})
}

func main() {
	m := http.NewServeMux()
	m.HandleFunc("POST /notas", crear)
	m.HandleFunc("GET /notas/{id}", leer)
	m.HandleFunc("PUT /notas/{id}", cambiar)
	m.HandleFunc("DELETE /notas/{id}", borrar)
	dir := ":8080"
	if v := os.Getenv("NOTAS_DIR"); v != "" {
		dir = v
	}
	log.SetFlags(0)
	log.Printf("notas-api en %s", dir)
	log.Fatal(http.ListenAndServe(dir, registrar(m)))
}
