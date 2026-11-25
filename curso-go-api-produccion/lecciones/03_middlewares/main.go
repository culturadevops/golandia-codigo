// Lección 3: middlewares. Tres trampas que no dan error.
//
//	go run ./lecciones/03_middlewares
package main

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"time"

	"tareas/internal/web"
)

func main() {
	estado()
	panico()
	plazo()
}

// ── 1. El estado que registra el log ──────────────────────────────────────────

type escritorIngenuo struct {
	http.ResponseWriter
	estado int
}

func (e *escritorIngenuo) WriteHeader(s int) { e.estado = s; e.ResponseWriter.WriteHeader(s) }

func registroIngenuo(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ew := &escritorIngenuo{ResponseWriter: w}
		next.ServeHTTP(ew, r)
		slog.Info("peticion", "ruta", r.URL.Path, "estado", ew.estado)
	})
}

func estado() {
	fmt.Println("1) un handler que solo hace Write (Go manda 200 solo):")
	h := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) { _, _ = io.WriteString(w, "ok") })
	var buf bytes.Buffer
	slog.SetDefault(slog.New(slog.NewTextHandler(&buf, &slog.HandlerOptions{ReplaceAttr: sinHora})))
	registroIngenuo(h).ServeHTTP(httptest.NewRecorder(), httptest.NewRequest("GET", "/tareas", nil))
	fmt.Print("  ingenuo:   ", buf.String())
	buf.Reset()
	web.Cadena(h, web.ConID(slog.Default()), web.Registro).ServeHTTP(httptest.NewRecorder(),
		httptest.NewRequest("GET", "/tareas", nil))
	fmt.Print("  plantilla: ", quitarID(buf.String()))
	fmt.Println()
}

// ── 2. Un pánico en un handler ───────────────────────────────────────────────

func panico() {
	fmt.Println("2) un handler que entra en pánico:")
	boom := http.HandlerFunc(func(http.ResponseWriter, *http.Request) { panic("índice fuera de rango") })

	var errlog bytes.Buffer
	srv := httptest.NewUnstartedServer(boom)
	srv.Config.ErrorLog = log.New(&errlog, "", 0)
	srv.Start()
	_, err := http.Get(srv.URL + "/tareas/1")
	srv.Close()
	fmt.Println("  sin Recuperar, el cliente recibe:", err)
	fmt.Println("  y el servidor escribe:", primeraLinea(errlog.String()))

	var buf bytes.Buffer
	lg := slog.New(slog.NewTextHandler(&buf, &slog.HandlerOptions{ReplaceAttr: sinPila}))
	srv = httptest.NewServer(web.Cadena(boom, web.ConID(lg), web.Registro, web.Recuperar))
	resp, err := http.Get(srv.URL + "/tareas/1")
	if err != nil {
		fmt.Println("  error inesperado:", err)
		return
	}
	cuerpo, _ := io.ReadAll(resp.Body)
	resp.Body.Close()
	srv.Close()
	fmt.Printf("  con Recuperar: %d %s", resp.StatusCode, cuerpo)
	for _, l := range strings.Split(strings.TrimSpace(buf.String()), "\n") {
		fmt.Println("    log:", l)
	}
	fmt.Println()
}

// ── 3. El plazo solo funciona si el handler lo mira ──────────────────────────

func plazo() {
	fmt.Println("3) plazo de 250 ms; el handler hace 10 pasos de 100 ms:")
	for _, mira := range []bool{false, true} {
		pasos := 0
		h := web.Manejador(func(w http.ResponseWriter, r *http.Request) error {
			for pasos < 10 {
				if mira {
					select {
					case <-r.Context().Done():
						return r.Context().Err()
					case <-time.After(100 * time.Millisecond):
					}
				} else {
					time.Sleep(100 * time.Millisecond) // trabajo que no mira el contexto
				}
				pasos++
			}
			web.JSON(w, 200, map[string]int{"pasos": pasos})
			return nil
		})
		rec := httptest.NewRecorder()
		t0 := time.Now()
		web.Plazo(250*time.Millisecond)(h).ServeHTTP(rec, httptest.NewRequest("GET", "/", nil))
		que := "ignora r.Context()"
		if mira {
			que = "mira r.Context()  "
		}
		fmt.Printf("  %s -> %d en %4d ms, %2d de 10 pasos hechos  %s",
			que, rec.Code, time.Since(t0).Round(50*time.Millisecond).Milliseconds(), pasos, rec.Body)
	}
}

func sinHora(_ []string, a slog.Attr) slog.Attr {
	if a.Key == slog.TimeKey {
		return slog.Attr{}
	}
	return a
}

func sinPila(g []string, a slog.Attr) slog.Attr {
	if a.Key == "pila" {
		return slog.String("pila", "(…"+fmt.Sprint(strings.Count(a.Value.String(), "\n"))+" líneas)")
	}
	if a.Key == "ms" {
		return slog.Attr{}
	}
	return sinHora(g, a)
}

func quitarID(s string) string {
	i := strings.Index(s, "request_id=")
	if i < 0 {
		return s
	}
	j := strings.IndexByte(s[i:], ' ')
	return s[:i] + "request_id=… " + s[i+j+1:]
}

func primeraLinea(s string) string {
	l, _, _ := strings.Cut(s, "\n")
	return l
}
