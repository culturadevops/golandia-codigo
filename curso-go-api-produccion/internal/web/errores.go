// Package web tiene lo que comparten todos los handlers: errores JSON, lectura del cuerpo y
// middlewares. No sabe nada de tareas ni de usuarios.
package web

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
)

// Error es un error que el cliente puede ver. Todo lo demás se registra y sale como 500.
type Error struct {
	Estado  int               `json:"-"`
	Codigo  string            `json:"codigo"`
	Mensaje string            `json:"mensaje"`
	Campos  map[string]string `json:"campos,omitempty"`
}

func (e *Error) Error() string { return e.Codigo + ": " + e.Mensaje }

// Errores con nombre, para no escribir el mismo código con dos grafías.
func NoEncontrado(m string) *Error  { return &Error{Estado: 404, Codigo: "no_encontrado", Mensaje: m} }
func Conflicto(m string) *Error     { return &Error{Estado: 409, Codigo: "conflicto", Mensaje: m} }
func NoAutenticado(m string) *Error { return &Error{Estado: 401, Codigo: "no_autenticado", Mensaje: m} }
func PeticionMala(m string) *Error  { return &Error{Estado: 400, Codigo: "peticion_invalida", Mensaje: m} }

// Validacion junta TODOS los campos mal en un solo error. Devuelve nil si no hay ninguno.
func Validacion(campos map[string]string) error {
	if len(campos) == 0 {
		return nil
	}
	return &Error{Estado: 422, Codigo: "validacion", Mensaje: "hay campos con problemas", Campos: campos}
}

type sobre struct {
	Error     *Error `json:"error"`
	RequestID string `json:"request_id,omitempty"`
}

// JSON escribe v con el estado dado.
func JSON(w http.ResponseWriter, estado int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(estado)
	if v != nil {
		_ = json.NewEncoder(w).Encode(v)
	}
}

// Fallo traduce cualquier error a la respuesta JSON de siempre.
func Fallo(w http.ResponseWriter, r *http.Request, err error) {
	var e *Error
	switch {
	case errors.As(err, &e):
	case errors.Is(err, context.DeadlineExceeded):
		e = &Error{Estado: 503, Codigo: "tiempo_agotado", Mensaje: "la petición tardó demasiado"}
	case errors.Is(err, context.Canceled):
		return // el cliente se fue: no hay a quién contestar
	default:
		Logger(r.Context()).Error("error interno", "err", err)
		e = &Error{Estado: 500, Codigo: "interno", Mensaje: "error interno"}
	}
	JSON(w, e.Estado, sobre{Error: e, RequestID: IDPeticion(r.Context())})
}

// Manejador es un handler que devuelve error en vez de escribirlo él.
type Manejador func(w http.ResponseWriter, r *http.Request) error

func (m Manejador) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if err := m(w, r); err != nil {
		Fallo(w, r, err)
	}
}

// Logger devuelve el logger de la petición (con su request_id) o el de por defecto.
func Logger(ctx context.Context) *slog.Logger {
	if l, ok := ctx.Value(claveLogger).(*slog.Logger); ok {
		return l
	}
	return slog.Default()
}
