// La MISMA API de pedidos, plana: un archivo, el handler valida y habla con Postgres directamente.
package main

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"os"
	"strconv"

	_ "github.com/jackc/pgx/v5/stdlib"
)

type pedido struct {
	ID       int64  `json:"id"`
	Cliente  string `json:"cliente"`
	Unidades int    `json:"unidades"`
	Precio   int    `json:"precio"`
	Total    int    `json:"total"`
}

func main() {
	db, err := sql.Open("pgx", os.Getenv("DATABASE_URL"))
	if err != nil {
		panic(err)
	}
	http.HandleFunc("POST /pedidos", func(w http.ResponseWriter, r *http.Request) {
		var p pedido
		if err := json.NewDecoder(r.Body).Decode(&p); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		switch {
		case p.Cliente == "":
			http.Error(w, "el pedido necesita un cliente", http.StatusUnprocessableEntity)
			return
		case p.Unidades < 1:
			http.Error(w, "el pedido necesita al menos una unidad", http.StatusUnprocessableEntity)
			return
		case p.Unidades*p.Precio > 1_000_000:
			http.Error(w, "el total pasa del límite de 1.000.000 de centavos", http.StatusUnprocessableEntity)
			return
		}
		err := db.QueryRowContext(r.Context(), "INSERT INTO pedidos (cliente, unidades, precio) VALUES ($1, $2, $3) RETURNING id",
			p.Cliente, p.Unidades, p.Precio).Scan(&p.ID)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		p.Total = p.Unidades * p.Precio
		json.NewEncoder(w).Encode(p)
	})
	http.HandleFunc("GET /pedidos/{id}", func(w http.ResponseWriter, r *http.Request) {
		p := pedido{}
		p.ID, _ = strconv.ParseInt(r.PathValue("id"), 10, 64)
		err := db.QueryRowContext(r.Context(), "SELECT cliente, unidades, precio FROM pedidos WHERE id = $1", p.ID).
			Scan(&p.Cliente, &p.Unidades, &p.Precio)
		if err == sql.ErrNoRows {
			http.Error(w, "pedido no encontrado", http.StatusNotFound)
			return
		}
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		p.Total = p.Unidades * p.Precio
		json.NewEncoder(w).Encode(p)
	})
	http.ListenAndServe(":8080", nil)
}
