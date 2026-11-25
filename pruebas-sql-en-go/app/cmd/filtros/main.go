// filtros: consulta dinámica. La ESTRUCTURA se arma en Go; los VALORES van siempre como $n.
package main

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"strings"

	"tienda/conexion"
)

type Filtro struct {
	Categoria *string
	MaxPrecio *int
	Buscar    *string
}

func construir(f Filtro) (string, []any) {
	q := "SELECT nombre FROM productos"
	var conds []string
	var args []any
	if f.Categoria != nil {
		args = append(args, *f.Categoria)
		conds = append(conds, fmt.Sprintf("categoria = $%d", len(args)))
	}
	if f.MaxPrecio != nil {
		args = append(args, *f.MaxPrecio)
		conds = append(conds, fmt.Sprintf("precio_centavos <= $%d", len(args)))
	}
	if f.Buscar != nil {
		args = append(args, "%"+*f.Buscar+"%")
		conds = append(conds, fmt.Sprintf("nombre ILIKE $%d", len(args)))
	}
	if len(conds) > 0 {
		q += " WHERE " + strings.Join(conds, " AND ")
	}
	return q + " ORDER BY id", args
}

// MAL: el valor del usuario pegado dentro del texto SQL.
func concatenado(cat string) string {
	return "SELECT nombre FROM productos WHERE categoria = '" + cat + "' ORDER BY id"
}

func main() {
	ctx := context.Background()
	db, err := conexion.Abrir(ctx)
	if err != nil {
		fmt.Fprintln(os.Stderr, "error al arrancar:", err)
		os.Exit(1)
	}
	defer db.Close()

	cafe, max, buscar := "cafe", 15000, "molido"
	casos := []struct {
		nombre string
		f      Filtro
	}{
		{"sin filtros", Filtro{}},
		{"categoria", Filtro{Categoria: &cafe}},
		{"categoria + precio + texto", Filtro{Categoria: &cafe, MaxPrecio: &max, Buscar: &buscar}},
	}
	for _, c := range casos {
		q, args := construir(c.f)
		fmt.Printf("== %s\n   sql:  %s\n   args: %v\n", c.nombre, q, args)
		n, _ := contar(ctx, db, q, args...)
		fmt.Printf("   filas: %d\n", n)
	}

	entrada := "' OR '1'='1"
	fmt.Printf("\n== entrada del usuario: %s\n", entrada)
	q := concatenado(entrada)
	n, err := contar(ctx, db, q)
	fmt.Printf("-- concatenado\n   sql:   %s\n   filas: %d  err: %v\n", q, n, err)
	q, args := construir(Filtro{Categoria: &entrada})
	n, err = contar(ctx, db, q, args...)
	fmt.Printf("-- placeholder\n   sql:   %s\n   args:  %v\n   filas: %d  err: %v\n", q, args, n, err)
}

func contar(ctx context.Context, db *sql.DB, q string, args ...any) (int, error) {
	rows, err := db.QueryContext(ctx, q, args...)
	if err != nil {
		return 0, err
	}
	defer rows.Close()
	n := 0
	for rows.Next() {
		n++
	}
	return n, rows.Err()
}
