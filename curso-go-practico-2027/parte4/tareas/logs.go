package main

import (
	"log/slog"
	"net/http"
	"os"
	"time"
)

// nuevoLogger escribe una línea JSON por evento en la salida estándar,
// que es lo que leen docker logs y cualquier recolector de logs.
func nuevoLogger(nivel slog.Level) *slog.Logger {
	return slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: nivel}))
}

// grabadora recuerda el código de estado que escribió el manejador.
type grabadora struct {
	http.ResponseWriter
	estado int
}

func (g *grabadora) WriteHeader(estado int) {
	g.estado = estado
	g.ResponseWriter.WriteHeader(estado)
}

// registrar envuelve la API: una línea de log por petición, con su estado y lo que tardó.
func registrar(log *slog.Logger, siguiente http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		inicio := time.Now()
		g := &grabadora{ResponseWriter: w, estado: http.StatusOK}
		siguiente.ServeHTTP(g, r)
		nivel := slog.LevelInfo
		if g.estado >= 500 {
			nivel = slog.LevelWarn
		}
		log.Log(r.Context(), nivel, "petición",
			"metodo", r.Method, "ruta", r.URL.Path, "patron", r.Pattern,
			"estado", g.estado, "ms", time.Since(inicio).Milliseconds())
	})
}
