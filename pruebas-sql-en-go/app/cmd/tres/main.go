// tres: la misma consulta de tres formas (string, go:embed y sqlc) contra la misma base.
package main

import (
	"context"
	"fmt"
	"os"

	"tienda/conexion"
	"tienda/consultas"
	"tienda/db"
)

func main() {
	ctx := context.Background()
	sqldb, err := conexion.Abrir(ctx)
	if err != nil {
		fmt.Fprintln(os.Stderr, "error al arrancar:", err)
		os.Exit(1)
	}
	defer sqldb.Close()
	fmt.Println("conectado: ping OK")

	a, err := consultas.PorCategoriaString(ctx, sqldb, "cafe", 3)
	mostrar("(a) string en Go", err, len(a), func(i int) string { return fmt.Sprintf("%-22s %6d", a[i].Nombre, a[i].Precio) })

	b, err := consultas.PorCategoriaEmbed(ctx, sqldb, "cafe", 3)
	mostrar("(b) .sql con go:embed", err, len(b), func(i int) string { return fmt.Sprintf("%-22s %6d", b[i].Nombre, b[i].Precio) })

	q := db.New(sqldb)
	c, err := q.PorCategoria(ctx, db.PorCategoriaParams{Categoria: "cafe", Limit: 3})
	mostrar("(c) sqlc", err, len(c), func(i int) string { return fmt.Sprintf("%-22s %6d", c[i].Nombre, c[i].PrecioCentavos) })
}

func mostrar(titulo string, err error, n int, fila func(int) string) {
	fmt.Println("==", titulo)
	if err != nil {
		fmt.Println("   ERROR:", err)
		return
	}
	for i := 0; i < n; i++ {
		fmt.Println("  ", fila(i))
	}
}
