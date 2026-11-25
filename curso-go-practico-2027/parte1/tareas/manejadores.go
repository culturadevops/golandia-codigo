package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strconv"
)

// Almacen es lo que la API necesita para guardar tareas.
// Hoy es un map en memoria; en la parte 2, Postgres.
type Almacen interface {
	Listar(ctx context.Context) ([]Tarea, error)
	Crear(ctx context.Context, titulo string) (Tarea, error)
	Obtener(ctx context.Context, id int64) (Tarea, error)
	Marcar(ctx context.Context, id int64, hecha bool) (Tarea, error)
	Borrar(ctx context.Context, id int64) error
}

type API struct {
	almacen Almacen
}

// Rutas usa los patrones de Go 1.22+: método, ruta y {comodines}.
func (a *API) Rutas() *http.ServeMux {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /tareas", a.listar)
	mux.HandleFunc("POST /tareas", a.crear)
	mux.HandleFunc("GET /tareas/{id}", a.obtener)
	mux.HandleFunc("PATCH /tareas/{id}", a.marcar)
	mux.HandleFunc("DELETE /tareas/{id}", a.borrar)
	return mux
}

func (a *API) listar(w http.ResponseWriter, r *http.Request) {
	lista, err := a.almacen.Listar(r.Context())
	if err != nil {
		a.errorAlmacen(w, err)
		return
	}
	responderJSON(w, http.StatusOK, lista)
}

func (a *API) crear(w http.ResponseWriter, r *http.Request) {
	var entrada struct {
		Titulo string `json:"titulo"`
	}
	if err := leerJSON(w, r, &entrada); err != nil {
		responderError(w, http.StatusBadRequest, err.Error())
		return
	}
	titulo, err := validarTitulo(entrada.Titulo)
	if err != nil {
		responderError(w, http.StatusUnprocessableEntity, err.Error())
		return
	}
	t, err := a.almacen.Crear(r.Context(), titulo)
	if err != nil {
		a.errorAlmacen(w, err)
		return
	}
	w.Header().Set("Location", fmt.Sprintf("/tareas/%d", t.ID))
	responderJSON(w, http.StatusCreated, t)
}

func (a *API) obtener(w http.ResponseWriter, r *http.Request) {
	id, err := leerID(r)
	if err != nil {
		responderError(w, http.StatusBadRequest, err.Error())
		return
	}
	t, err := a.almacen.Obtener(r.Context(), id)
	if err != nil {
		a.errorAlmacen(w, err)
		return
	}
	responderJSON(w, http.StatusOK, t)
}

func (a *API) marcar(w http.ResponseWriter, r *http.Request) {
	id, err := leerID(r)
	if err != nil {
		responderError(w, http.StatusBadRequest, err.Error())
		return
	}
	var entrada struct {
		Hecha *bool `json:"hecha"`
	}
	if err := leerJSON(w, r, &entrada); err != nil {
		responderError(w, http.StatusBadRequest, err.Error())
		return
	}
	if entrada.Hecha == nil {
		responderError(w, http.StatusUnprocessableEntity, "falta el campo hecha")
		return
	}
	t, err := a.almacen.Marcar(r.Context(), id, *entrada.Hecha)
	if err != nil {
		a.errorAlmacen(w, err)
		return
	}
	responderJSON(w, http.StatusOK, t)
}

func (a *API) borrar(w http.ResponseWriter, r *http.Request) {
	id, err := leerID(r)
	if err != nil {
		responderError(w, http.StatusBadRequest, err.Error())
		return
	}
	if err := a.almacen.Borrar(r.Context(), id); err != nil {
		a.errorAlmacen(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// leerID saca el {id} de la ruta y comprueba que sea un número positivo.
func leerID(r *http.Request) (int64, error) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil || id < 1 {
		return 0, fmt.Errorf("id inválido: %q", r.PathValue("id"))
	}
	return id, nil
}

// leerJSON limita el cuerpo a 1 MiB y rechaza campos que no conoce.
func leerJSON(w http.ResponseWriter, r *http.Request, destino any) error {
	r.Body = http.MaxBytesReader(w, r.Body, 1<<20)
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	if err := dec.Decode(destino); err != nil {
		return fmt.Errorf("JSON inválido: %w", err)
	}
	return nil
}

func responderJSON(w http.ResponseWriter, estado int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(estado)
	json.NewEncoder(w).Encode(v)
}

func responderError(w http.ResponseWriter, estado int, mensaje string) {
	responderJSON(w, estado, map[string]string{"error": mensaje})
}

func (a *API) errorAlmacen(w http.ResponseWriter, err error) {
	if errors.Is(err, ErrNoEncontrada) {
		responderError(w, http.StatusNotFound, err.Error())
		return
	}
	log.Printf("error del almacén: %v", err)
	responderError(w, http.StatusInternalServerError, "error interno")
}
