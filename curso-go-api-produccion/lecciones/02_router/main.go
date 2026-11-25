// Lección 2: el router de net/http (patrones de Go 1.22+).
//
//	go run ./lecciones/02_router
//	GODEBUG=httpmuxgo121=1 go run ./lecciones/02_router   el mux viejo: los patrones no casan
package main

import (
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"

	"tareas/internal/cuentas"
	"tareas/internal/servidor"
	"tareas/internal/tareas"
)

func texto(w http.ResponseWriter, s string) { _, _ = io.WriteString(w, s) }

// viejo: el estilo de antes de 1.22. Ni método ni parámetros: todo a mano.
func viejo() http.Handler {
	m := http.NewServeMux()
	m.HandleFunc("/", func(w http.ResponseWriter, _ *http.Request) { texto(w, "inicio") })
	m.HandleFunc("/tareas/", func(w http.ResponseWriter, r *http.Request) {
		texto(w, "tarea "+strings.TrimPrefix(r.URL.Path, "/tareas/"))
	})
	return m
}

// nuevo: método y comodines en el patrón; {$} es «exactamente /».
func nuevo() *http.ServeMux {
	m := http.NewServeMux()
	m.HandleFunc("GET /{$}", func(w http.ResponseWriter, _ *http.Request) { texto(w, "inicio") })
	m.HandleFunc("GET /tareas/{id}", func(w http.ResponseWriter, r *http.Request) {
		texto(w, "tarea "+r.PathValue("id"))
	})
	m.HandleFunc("DELETE /tareas/{id}", func(w http.ResponseWriter, r *http.Request) {
		texto(w, "borrada "+r.PathValue("id"))
	})
	return m
}

// conComodin: lo de nuevo() más un "/" para devolver 404 en JSON. Se come los 405.
func conComodin() http.Handler {
	m := nuevo()
	m.HandleFunc("/", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(404)
		texto(w, `{"error":"no existe"}`)
	})
	return m
}

func probar(nombre string, h http.Handler, peticiones ...string) {
	fmt.Println(nombre)
	for _, p := range peticiones {
		metodo, ruta, _ := strings.Cut(p, " ")
		rec := httptest.NewRecorder()
		h.ServeHTTP(rec, httptest.NewRequest(metodo, ruta, nil))
		cuerpo := strings.TrimSpace(rec.Body.String())
		extra := ""
		if a := rec.Header().Get("Allow"); a != "" {
			extra = "  Allow: " + a
		}
		fmt.Printf("  %-20s -> %d  %s%s\n", p, rec.Code, cuerpo, extra)
	}
	fmt.Println()
}

func main() {
	if os.Getenv("GODEBUG") != "" {
		fmt.Println("GODEBUG=" + os.Getenv("GODEBUG"))
		probar("patrones de 1.22 con el mux viejo:", nuevo(), "GET /tareas/7", "GET /")
		return
	}
	probar("antes de 1.22 (\"/\" y \"/tareas/\"):", viejo(),
		"GET /tareas/7", "DELETE /tareas/7", "GET /tarea", "GET /no-existe")
	probar("Go 1.22+ (\"GET /{$}\", \"GET /tareas/{id}\"):", nuevo(),
		"GET /tareas/7", "POST /tareas/7", "GET /tarea", "GET /tareas/7/x")
	probar("lo mismo + un \"/\" para el 404 en JSON:", conComodin(), "POST /tareas/7")

	// El router de la plantilla, sin base de datos: solo rutas que no la tocan.
	h := servidor.Rutas(servidor.Deps{
		Log:     slog.New(slog.NewJSONHandler(io.Discard, nil)),
		Cuentas: cuentas.Nuevo(nil, 10),
		Tareas:  tareas.Postgres{},
		Salud:   &servidor.Salud{},
	})
	probar("el de la plantilla (ErroresDelMux):", h, "DELETE /salud/vivo", "GET /no-existe")
}
