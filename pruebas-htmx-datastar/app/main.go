// Demo del vídeo «¿HTMX o Datastar? Con Go» (pedido en un comentario de «Go + HTMX», 2026-08-08: «una comparación entre HTMX y
// Datastar»). La MISMA app dos veces: una búsqueda en vivo y un reloj que se actualiza solo.
//   /htmx      → htmx 2.0.11: el servidor devuelve fragmentos de HTML; el reloj pregunta cada segundo (polling).
//   /datastar  → Datastar 1.0.4 + datastar-go: el servidor responde con eventos SSE; el reloj es UNA conexión que el servidor empuja.
// Las dos librerías se sirven desde aquí (static/), sin CDN, para medir lo que de verdad se descarga.
package main

import (
	_ "embed"
	"fmt"
	"html"
	"net/http"
	"strings"
	"time"

	"github.com/starfederation/datastar-go/datastar"
)

//go:embed static/htmx.min.js
var htmxJS []byte

//go:embed static/datastar.js
var datastarJS []byte

var lenguajes = []string{"Go", "Rust", "Zig", "Java", "Kotlin", "C#", "Elixir", "Python", "TypeScript", "Ruby", "PHP", "Swift",
	"Scala", "Haskell", "Gleam", "Odin", "Dart", "Lua", "Julia", "Clojure"}

func buscar(q string) string {
	q = strings.ToLower(strings.TrimSpace(q))
	var b strings.Builder
	b.WriteString(`<ul id="resultados">`)
	for _, l := range lenguajes {
		if q == "" || strings.Contains(strings.ToLower(l), q) {
			b.WriteString("<li>" + html.EscapeString(l) + "</li>")
		}
	}
	b.WriteString("</ul>")
	return b.String()
}

const paginaHTMX = `<!doctype html><html lang="es"><head><meta charset="utf-8"><title>HTMX</title>
<script src="/static/htmx.min.js"></script></head><body>
<h1>Con HTMX</h1>
<input name="q" placeholder="Busca un lenguaje" hx-get="/htmx/buscar" hx-trigger="input changed delay:300ms" hx-target="#resultados" hx-swap="outerHTML">
<ul id="resultados"></ul>
<p>Hora del servidor: <span hx-get="/htmx/hora" hx-trigger="load, every 1s">…</span></p>
</body></html>`

const paginaDatastar = `<!doctype html><html lang="es"><head><meta charset="utf-8"><title>Datastar</title>
<script type="module" src="/static/datastar.js"></script></head><body data-signals:q="''">
<h1>Con Datastar</h1>
<input data-bind:q placeholder="Busca un lenguaje" data-on:input__debounce.300ms="@get('/datastar/buscar')">
<ul id="resultados"></ul>
<p>Hora del servidor: <span id="hora" data-init="@get('/datastar/hora')">…</span></p>
</body></html>`

func main() {
	http.HandleFunc("GET /static/htmx.min.js", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/javascript")
		w.Write(htmxJS)
	})
	http.HandleFunc("GET /static/datastar.js", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/javascript")
		w.Write(datastarJS)
	})

	// ── HTMX: HTML de ida y vuelta ──
	http.HandleFunc("GET /htmx", func(w http.ResponseWriter, r *http.Request) { fmt.Fprint(w, paginaHTMX) })
	http.HandleFunc("GET /htmx/buscar", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, buscar(r.URL.Query().Get("q")))
	})
	http.HandleFunc("GET /htmx/hora", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, time.Now().Format("15:04:05"))
	})

	// ── Datastar: señales de ida, eventos SSE de vuelta ──
	http.HandleFunc("GET /datastar", func(w http.ResponseWriter, r *http.Request) { fmt.Fprint(w, paginaDatastar) })
	http.HandleFunc("GET /datastar/buscar", func(w http.ResponseWriter, r *http.Request) {
		var s struct {
			Q string `json:"q"`
		}
		if err := datastar.ReadSignals(r, &s); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		datastar.NewSSE(w, r).PatchElements(buscar(s.Q))
	})
	http.HandleFunc("GET /datastar/hora", func(w http.ResponseWriter, r *http.Request) {
		sse := datastar.NewSSE(w, r)
		for {
			if err := sse.PatchElements(`<span id="hora">` + time.Now().Format("15:04:05") + `</span>`); err != nil {
				return // el navegador cerró la conexión
			}
			select {
			case <-r.Context().Done():
				return
			case <-time.After(time.Second):
			}
		}
	})

	fmt.Println("escuchando en :8080")
	http.ListenAndServe(":8080", nil)
}
