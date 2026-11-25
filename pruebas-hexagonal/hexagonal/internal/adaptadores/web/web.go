// Package web es el ADAPTADOR de entrada: traduce HTTP a llamadas al servicio. El dominio no sabe que existe HTTP.
package web

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"tienda/internal/pedidos"
)

type pedidoJSON struct {
	ID       int64  `json:"id"`
	Cliente  string `json:"cliente"`
	Unidades int    `json:"unidades"`
	Precio   int    `json:"precio"`
	Total    int    `json:"total"`
}

func aJSON(p pedidos.Pedido) pedidoJSON {
	return pedidoJSON{p.ID, p.Cliente, p.Unidades, p.Precio, p.Total()}
}

func Rutas(s *pedidos.Servicio) *http.ServeMux {
	m := http.NewServeMux()
	m.HandleFunc("POST /pedidos", func(w http.ResponseWriter, r *http.Request) {
		var in pedidoJSON
		if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		p, err := s.Crear(r.Context(), in.Cliente, in.Unidades, in.Precio)
		if err != nil {
			http.Error(w, err.Error(), http.StatusUnprocessableEntity)
			return
		}
		json.NewEncoder(w).Encode(aJSON(p))
	})
	m.HandleFunc("GET /pedidos/{id}", func(w http.ResponseWriter, r *http.Request) {
		id, _ := strconv.ParseInt(r.PathValue("id"), 10, 64)
		p, err := s.Buscar(r.Context(), id)
		if errors.Is(err, pedidos.ErrNoEncontrado) {
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		json.NewEncoder(w).Encode(aJSON(p))
	})
	return m
}
