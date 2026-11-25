// Package api expone el servicio por HTTP.
//
//	POST /pedidos       {"cliente":"ana","items":[{"sku":"TAZA","cantidad":2,"precio_centavos":15000}]}
//	GET  /pedidos/{id}
package api

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"curso-testing/tienda"
)

func Rutas(s *tienda.Servicio) *http.ServeMux {
	mux := http.NewServeMux()
	mux.HandleFunc("POST /pedidos", crear(s))
	mux.HandleFunc("GET /pedidos/{id}", ver(s))
	return mux
}

func crear(s *tienda.Servicio) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var in struct {
			Cliente string        `json:"cliente"`
			Items   []tienda.Item `json:"items"`
		}
		if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, 1<<20)).Decode(&in); err != nil {
			responder(w, http.StatusBadRequest, map[string]string{"error": "JSON inválido"})
			return
		}
		if in.Cliente == "" {
			responder(w, http.StatusBadRequest, map[string]string{"error": "falta el cliente"})
			return
		}
		p, err := s.Crear(r.Context(), in.Cliente, in.Items)
		switch {
		case errors.Is(err, tienda.ErrPedidoVacio), errors.Is(err, tienda.ErrCantidad):
			responder(w, http.StatusUnprocessableEntity, map[string]string{"error": err.Error()})
		case err != nil:
			responder(w, http.StatusBadGateway, map[string]string{"error": "no se pudo cobrar"})
		default:
			responder(w, http.StatusCreated, p)
		}
	}
}

func ver(s *tienda.Servicio) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
		if err != nil {
			responder(w, http.StatusBadRequest, map[string]string{"error": "id inválido"})
			return
		}
		p, err := s.Repo.Buscar(r.Context(), id)
		if errors.Is(err, tienda.ErrNoExiste) {
			responder(w, http.StatusNotFound, map[string]string{"error": "no existe"})
			return
		}
		if err != nil {
			responder(w, http.StatusInternalServerError, map[string]string{"error": "error interno"})
			return
		}
		responder(w, http.StatusOK, p)
	}
}

func responder(w http.ResponseWriter, codigo int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(codigo)
	_ = json.NewEncoder(w).Encode(v)
}
