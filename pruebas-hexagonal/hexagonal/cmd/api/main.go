// El cableado: el ÚNICO sitio que sabe qué adaptador va en cada puerto. Cambiar de Postgres a memoria es cambiar una línea aquí.
package main

import (
	"database/sql"
	"net/http"
	"os"

	_ "github.com/jackc/pgx/v5/stdlib"

	"tienda/internal/adaptadores/memoria"
	"tienda/internal/adaptadores/postgres"
	"tienda/internal/adaptadores/web"
	"tienda/internal/pedidos"
)

func main() {
	var repo pedidos.Repositorio
	if dsn := os.Getenv("DATABASE_URL"); dsn != "" {
		db, err := sql.Open("pgx", dsn)
		if err != nil {
			panic(err)
		}
		repo = postgres.Nuevo(db)
	} else {
		repo = memoria.Nuevo()
	}
	http.ListenAndServe(":8080", web.Rutas(pedidos.NuevoServicio(repo)))
}
