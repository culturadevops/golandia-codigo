// Package consultas: la MISMA consulta escrita de dos formas a mano (la tercera la genera sqlc en ../db).
package consultas

import (
	"context"
	"database/sql"
	_ "embed"
)

type Producto struct {
	ID       int64
	Nombre   string
	Precio   int32 // centavos
}

// (a) El SQL como string dentro del código Go, con placeholders $1 y $2.
const porCategoria = `SELECT id, nombre, precio_centavos
FROM productos
WHERE categoria = $1
ORDER BY precio_centavos
LIMIT $2`

func PorCategoriaString(ctx context.Context, db *sql.DB, cat string, limite int32) ([]Producto, error) {
	return leer(ctx, db, porCategoria, cat, limite)
}

// (b) El SQL en un archivo .sql, metido en el binario con go:embed.
//
//go:embed sql/por_categoria.sql
var porCategoriaEmbed string

func PorCategoriaEmbed(ctx context.Context, db *sql.DB, cat string, limite int32) ([]Producto, error) {
	return leer(ctx, db, porCategoriaEmbed, cat, limite)
}

func leer(ctx context.Context, db *sql.DB, q string, args ...any) ([]Producto, error) {
	rows, err := db.QueryContext(ctx, q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var ps []Producto
	for rows.Next() {
		var p Producto
		if err := rows.Scan(&p.ID, &p.Nombre, &p.Precio); err != nil {
			return nil, err
		}
		ps = append(ps, p)
	}
	return ps, rows.Err()
}
