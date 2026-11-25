package web

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"log/slog"
	"net/http"
	"runtime/debug"
	"strings"
	"time"
)

type clave int

const (
	claveID clave = iota
	claveLogger
)

// Cadena aplica los middlewares en el orden en que se leen: el primero es el de fuera.
func Cadena(h http.Handler, mws ...func(http.Handler) http.Handler) http.Handler {
	for i := len(mws) - 1; i >= 0; i-- {
		h = mws[i](h)
	}
	return h
}

// IDPeticion devuelve el request id de la petición, o "".
func IDPeticion(ctx context.Context) string {
	id, _ := ctx.Value(claveID).(string)
	return id
}

// ConID pone un request id (el del cliente si es razonable) y un logger que lo lleva.
// Va el PRIMERO: así hasta el log de un pánico sale con su id.
func ConID(log *slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			id := r.Header.Get("X-Request-Id")
			if !idValido(id) {
				id = nuevoID()
			}
			w.Header().Set("X-Request-Id", id)
			ctx := context.WithValue(r.Context(), claveID, id)
			ctx = context.WithValue(ctx, claveLogger, log.With("request_id", id))
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// idValido rechaza ids largos o con caracteres raros: el id acaba en los logs.
func idValido(id string) bool {
	if id == "" || len(id) > 64 {
		return false
	}
	for _, c := range id {
		if !(c == '-' || c >= '0' && c <= '9' || c >= 'a' && c <= 'z' || c >= 'A' && c <= 'Z') {
			return false
		}
	}
	return true
}

func nuevoID() string {
	b := make([]byte, 8)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}

// Registro escribe una línea por petición, con el estado REAL que se mandó.
func Registro(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t0 := time.Now()
		ew := &escritor{ResponseWriter: w}
		next.ServeHTTP(ew, r)
		nivel := slog.LevelInfo
		if strings.HasPrefix(r.URL.Path, "/salud/") && ew.Estado() == http.StatusOK {
			nivel = slog.LevelDebug // la sonda pasa cada 5 s: no llena el log si va bien
		}
		Logger(r.Context()).Log(r.Context(), nivel, "peticion", "metodo", r.Method, "ruta", r.URL.Path,
			"estado", ew.Estado(), "bytes", ew.bytes, "ms", time.Since(t0).Milliseconds())
	})
}

// escritor recuerda el estado. Si el handler nunca llama a WriteHeader, Go manda 200 al
// primer Write: el envoltorio tiene que saberlo, o el log dirá 0.
type escritor struct {
	http.ResponseWriter
	estado int
	bytes  int
}

func (e *escritor) WriteHeader(s int) {
	if e.estado == 0 {
		e.estado = s
	}
	e.ResponseWriter.WriteHeader(s)
}

func (e *escritor) Write(b []byte) (int, error) {
	if e.estado == 0 {
		e.estado = http.StatusOK
	}
	n, err := e.ResponseWriter.Write(b)
	e.bytes += n
	return n, err
}

func (e *escritor) Estado() int {
	if e.estado == 0 {
		return http.StatusOK
	}
	return e.estado
}

// Unwrap deja a http.ResponseController llegar al writer de verdad (Flush, plazos).
func (e *escritor) Unwrap() http.ResponseWriter { return e.ResponseWriter }

// Recuperar convierte un pánico en un 500 JSON y un log con la pila. Sin él, net/http
// también se recupera, pero corta la conexión: el cliente ve un EOF, no un error.
func Recuperar(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			v := recover()
			if v == nil {
				return
			}
			if v == http.ErrAbortHandler {
				panic(v) // el aborto a propósito se respeta
			}
			Logger(r.Context()).Error("panico", "valor", v, "pila", string(debug.Stack()))
			JSON(w, 500, sobre{Error: &Error{Codigo: "interno", Mensaje: "error interno"},
				RequestID: IDPeticion(r.Context())})
		}()
		next.ServeHTTP(w, r)
	})
}

// Plazo le pone fecha límite al contexto de la petición. Solo sirve si los handlers pasan
// r.Context() a lo que hacen (la base de datos, otras llamadas): Go no interrumpe a nadie.
func Plazo(d time.Duration) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		if d <= 0 {
			return next // un plazo de 0 vencería antes de empezar
		}
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx, cancel := context.WithTimeout(r.Context(), d)
			defer cancel()
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
