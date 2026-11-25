// Squirrel en Go: consultas SQL dinámicas sin concatenar strings (golandia 881).
// El mismo buscador de productos (filtros opcionales, orden y paginación) de cuatro formas, ejecutado contra Postgres:
//
//	A · fmt.Sprintf (insegura: para ver la inyección)
//	B · a mano, con placeholders $1, $2…
//	C · Squirrel (github.com/Masterminds/squirrel) con PlaceholderFormat(sq.Dollar)
//	D · sqlc, con sqlc.narg para los filtros opcionales (código generado en ./bd)
//
// Imprime el SQL que genera cada forma, sus argumentos y lo que devuelve Postgres. ⛔ Las inyecciones se ejecutan SOLO contra
// el Postgres desechable del banco (contenedor con --rm y datos en tmpfs).
package main

import (
	"context"
	"fmt"
	"os"
	"slices"
	"strings"

	"golandia/squirrel/bd"

	sq "github.com/Masterminds/squirrel"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Filtro: lo que llega del buscador. Valor cero = «no filtrar por esto».
type Filtro struct {
	Categoria string
	PrecioMax float64
	Busca     string
	IDs       []int64
	Orden     string // lo que manda el usuario: "precio", "-precio", "nombre"…
	Pagina    int
	PorPagina int
}

// ordenes: la lista cerrada. El ORDER BY no admite placeholders: o es una de estas, o es "id".
var ordenes = map[string]string{"precio": "precio, id", "-precio": "precio DESC, id", "nombre": "nombre, id"}

func ordenSeguro(o string) string {
	if v, ok := ordenes[o]; ok {
		return v
	}
	return "id"
}

// A · fmt.Sprintf: el texto del usuario se pega dentro del SQL.
func conSprintf(f Filtro) string {
	q := "SELECT id, nombre, precio FROM productos WHERE 1=1"
	if f.Categoria != "" {
		q += fmt.Sprintf(" AND categoria = '%s'", f.Categoria)
	}
	if f.PrecioMax > 0 {
		q += fmt.Sprintf(" AND precio <= %v", f.PrecioMax)
	}
	if f.Busca != "" {
		q += fmt.Sprintf(" AND nombre ILIKE '%%%s%%'", f.Busca)
	}
	q += " ORDER BY " + ordenSeguro(f.Orden)
	q += fmt.Sprintf(" LIMIT %d OFFSET %d", f.PorPagina, f.Pagina*f.PorPagina)
	return q
}

// B · a mano: el SQL se arma en Go, los valores van aparte y se cuentan los $n.
func aMano(f Filtro) (string, []any) {
	var conds []string
	var args []any
	agregar := func(cond string, v any) {
		args = append(args, v)
		conds = append(conds, fmt.Sprintf(cond, len(args)))
	}
	if f.Categoria != "" {
		agregar("categoria = $%d", f.Categoria)
	}
	if f.PrecioMax > 0 {
		agregar("precio <= $%d", f.PrecioMax)
	}
	if f.Busca != "" {
		agregar("nombre ILIKE $%d", "%"+f.Busca+"%")
	}
	if len(f.IDs) > 0 {
		agregar("id = ANY($%d)", f.IDs)
	}
	q := "SELECT id, nombre, precio FROM productos"
	if len(conds) > 0 {
		q += " WHERE " + strings.Join(conds, " AND ")
	}
	q += " ORDER BY " + ordenSeguro(f.Orden)
	args = append(args, f.PorPagina)
	q += fmt.Sprintf(" LIMIT $%d", len(args))
	args = append(args, f.Pagina*f.PorPagina)
	q += fmt.Sprintf(" OFFSET $%d", len(args))
	return q, args
}

// C · Squirrel. psql: el constructor con los placeholders de Postgres ($1). Sin esto, Squirrel escribe «?».
var psql = sq.StatementBuilder.PlaceholderFormat(sq.Dollar)

func conSquirrel(f Filtro) (string, []any, error) {
	q := psql.Select("id", "nombre", "precio").From("productos")
	if f.Categoria != "" {
		q = q.Where(sq.Eq{"categoria": f.Categoria})
	}
	if f.PrecioMax > 0 {
		q = q.Where(sq.LtOrEq{"precio": f.PrecioMax})
	}
	if f.Busca != "" {
		q = q.Where(sq.ILike{"nombre": "%" + f.Busca + "%"})
	}
	if len(f.IDs) > 0 {
		q = q.Where(sq.Eq{"id": f.IDs})
	}
	q = q.OrderBy(ordenSeguro(f.Orden)).Limit(uint64(f.PorPagina)).Offset(uint64(f.Pagina * f.PorPagina))
	return q.ToSql()
}

var ctx = context.Background()

func ids(pool *pgxpool.Pool, sql string, args ...any) ([]int64, error) {
	rows, err := pool.Query(ctx, sql, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []int64
	for rows.Next() {
		var id int64
		var nombre string
		var precio pgtype.Numeric
		if err := rows.Scan(&id, &nombre, &precio); err != nil {
			return nil, err
		}
		out = append(out, id)
	}
	return out, rows.Err()
}

func resumen(v []int64) string {
	if len(v) > 5 {
		return fmt.Sprintf("%d filas, primeros ids %v…", len(v), v[:5])
	}
	return fmt.Sprintf("%d filas, ids %v", len(v), v)
}

func titulo(s string) { fmt.Printf("\n## %s\n\n", s) }

func cuenta(pool *pgxpool.Pool, tabla string) int {
	var n int
	if err := pool.QueryRow(ctx, "SELECT count(*) FROM "+tabla).Scan(&n); err != nil {
		panic(err)
	}
	return n
}

func main() {
	pool, err := pgxpool.New(ctx, os.Getenv("DATABASE_URL"))
	if err != nil {
		panic(err)
	}
	defer pool.Close()
	var ver string
	pool.QueryRow(ctx, "SHOW server_version").Scan(&ver)
	fmt.Printf("Postgres %s · productos: %d filas\n", ver, cuenta(pool, "productos"))

	// ── 1 · Las tres formas, con entradas normales: ¿devuelven lo mismo? ──
	titulo("1 · El mismo buscador de tres formas, con entradas normales")
	casos := []struct {
		nombre string
		f      Filtro
	}{
		{"sin filtros", Filtro{PorPagina: 5}},
		{"categoría", Filtro{Categoria: "cafe", PorPagina: 5}},
		{"categoría + precio + texto + orden + página 2", Filtro{Categoria: "cafe", PrecioMax: 10, Busca: "99", Orden: "-precio",
			Pagina: 1, PorPagina: 5}},
	}
	for _, c := range casos {
		fmt.Printf("### %s\n\n", c.nombre)
		a := conSprintf(c.f)
		b, bargs := aMano(c.f)
		s, sargs, err := conSquirrel(c.f)
		if err != nil {
			panic(err)
		}
		ra, ea := ids(pool, a)
		rb, eb := ids(pool, b, bargs...)
		rs, es := ids(pool, s, sargs...)
		if ea != nil || eb != nil || es != nil {
			panic(fmt.Sprint(ea, eb, es))
		}
		fmt.Printf("A Sprintf:  %s\n            → %s\n", a, resumen(ra))
		fmt.Printf("B a mano:   %s\n            args %v → %s\n", b, bargs, resumen(rb))
		fmt.Printf("C Squirrel: %s\n            args %v → %s\n", s, sargs, resumen(rs))
		if !slices.Equal(ra, rb) || !slices.Equal(rb, rs) {
			fmt.Println("⛔ DISTINTO: no se compara")
			os.Exit(1)
		}
		fmt.Printf("¿lo mismo? sí, las tres\n\n")
	}

	// ── 2 · La misma entrada maliciosa en el campo categoría ──
	titulo("2 · Entrada maliciosa en «categoría»: nada' OR '1'='1")
	mal := Filtro{Categoria: "nada' OR '1'='1", PorPagina: 100000}
	a := conSprintf(mal)
	ra, err := ids(pool, a)
	fmt.Printf("A Sprintf:  %s\n            → %d filas (err %v)\n", a, len(ra), err)
	b, bargs := aMano(mal)
	rb, err := ids(pool, b, bargs...)
	fmt.Printf("B a mano:   %s\n            args %q → %d filas (err %v)\n", b, bargs[0], len(rb), err)
	s, sargs, _ := conSquirrel(mal)
	rs, err := ids(pool, s, sargs...)
	fmt.Printf("C Squirrel: %s\n            args %q → %d filas (err %v)\n", s, sargs[0], len(rs), err)

	// ── 3 · Inyección apilada: depende del protocolo ──
	titulo("3 · Inyección apilada con Sprintf: x'; DELETE FROM notas; --")
	apilada := conSprintf(Filtro{Categoria: "x'; DELETE FROM notas; --", PorPagina: 5})
	fmt.Println("SQL:", apilada)
	fmt.Printf("notas antes: %d\n", cuenta(pool, "notas"))
	_, err = ids(pool, apilada)
	fmt.Printf("pgx por defecto (protocolo extendido): err = %v\n", err)
	fmt.Printf("notas después: %d\n", cuenta(pool, "notas"))
	rows, err := pool.Query(ctx, apilada, pgx.QueryExecModeSimpleProtocol)
	if err == nil {
		for rows.Next() {
		}
		err = rows.Err()
		rows.Close()
	}
	fmt.Printf("pgx con QueryExecModeSimpleProtocol: err = %v\n", err)
	fmt.Printf("notas después: %d\n", cuenta(pool, "notas"))

	// ── 4 · Trampa: Squirrel sin PlaceholderFormat ──
	titulo("4 · Squirrel sin PlaceholderFormat(sq.Dollar), contra Postgres")
	q4, a4, _ := sq.Select("id", "nombre", "precio").From("productos").Where(sq.Eq{"categoria": "cafe"}).Limit(3).ToSql()
	fmt.Printf("SQL: %s   args %v\n", q4, a4)
	_, err = ids(pool, q4, a4...)
	fmt.Printf("err = %v\n", err)
	q4b, _, _ := psql.Select("id").From("productos").Where(sq.Eq{"categoria": "cafe"}).Limit(3).ToSql()
	fmt.Printf("con psql: %s\n", q4b)

	// ── 5 · Trampa: OrderBy con lo que manda el usuario ──
	titulo("5 · Squirrel .OrderBy(entrada del usuario): el ORDER BY no lleva placeholder")
	for _, letra := range []string{"a", "s"} {
		entrada := fmt.Sprintf("(CASE WHEN (SELECT substr(clave,1,1) FROM usuarios WHERE nombre='admin')='%s' THEN precio END), id", letra)
		qs, as, _ := psql.Select("id", "nombre", "precio").From("productos").Where(sq.Eq{"categoria": "cafe"}).OrderBy(entrada).Limit(5).ToSql()
		r, err := ids(pool, qs, as...)
		fmt.Printf("orden = %q\n  SQL: %s\n  → %s (err %v)\n", entrada, qs, resumen(r), err)
	}
	r5, _ := ids(pool, "SELECT id, nombre, precio FROM productos WHERE categoria='cafe' ORDER BY precio, id LIMIT 5")
	fmt.Printf("referencia, ORDER BY precio, id → %s\n", resumen(r5))
	fmt.Printf("con la lista cerrada, la misma entrada → ORDER BY %s\n", ordenSeguro("(CASE WHEN ...)"))

	// ── 6 · Listas: IN ──
	titulo("6 · Listas: sq.Eq con un slice")
	for _, l := range [][]int64{{3, 7, 42}, {}} {
		q6, a6, _ := psql.Select("id").From("productos").Where(sq.Eq{"id": l}).ToSql()
		fmt.Printf("sq.Eq{\"id\": %v} → %s   args %v\n", l, q6, a6)
	}
	q6m, a6m := aMano(Filtro{IDs: []int64{3, 7, 42}, PorPagina: 10})
	r6, _ := ids(pool, q6m, a6m...)
	fmt.Printf("a mano: %s   args %v → %s\n", q6m, a6m, resumen(r6))

	// ── 7 · Paginación: OFFSET y por cursor ──
	titulo("7 · Paginación: OFFSET contra cursor (keyset), página 3 de 5 en 5")
	po, pa, _ := psql.Select("id", "nombre", "precio").From("productos").Where(sq.Eq{"categoria": "te"}).OrderBy("id").Limit(5).Offset(10).ToSql()
	ro, _ := ids(pool, po, pa...)
	fmt.Printf("OFFSET: %s   args %v → %s\n", po, pa, resumen(ro))
	ultimo := int64(0)
	var rk []int64
	for i := 0; i < 3; i++ {
		pk, ka, _ := psql.Select("id", "nombre", "precio").From("productos").Where(sq.Eq{"categoria": "te"}).Where(sq.Gt{"id": ultimo}).OrderBy("id").Limit(5).ToSql()
		rk, _ = ids(pool, pk, ka...)
		if i == 2 {
			fmt.Printf("cursor: %s   args %v → %s\n", pk, ka, resumen(rk))
		}
		ultimo = rk[len(rk)-1]
	}
	fmt.Printf("¿la misma página? %v\n", slices.Equal(ro, rk))

	// ── 8 · sqlc: filtros opcionales con sqlc.narg ──
	titulo("8 · sqlc: la consulta fija con sqlc.narg")
	fmt.Println("SQL generado (bd/query.sql.go):")
	fmt.Println(bd.SqlBuscar)
	qs := bd.New(pool)
	todos, err := qs.Buscar(ctx, bd.BuscarParams{Limite: 100000})
	fmt.Printf("sin filtros → %d filas (err %v)\n", len(todos), err)
	cafe, err := qs.Buscar(ctx, bd.BuscarParams{Categoria: pgtype.Text{String: "cafe", Valid: true}, Limite: 100000})
	fmt.Printf("categoria=cafe → %d filas (err %v)\n", len(cafe), err)
	s8, a8, _ := conSquirrel(Filtro{Categoria: "cafe", PorPagina: 100000})
	r8, _ := ids(pool, s8, a8...)
	fmt.Printf("Squirrel, categoria=cafe → %d filas\n", len(r8))

	titulo("9 · sqlc: el orden como parámetro")
	fmt.Println(bd.SqlBuscarOrden)
	ro9, err := qs.BuscarOrden(ctx, bd.BuscarOrdenParams{Categoria: "cafe", Orden: "precio"})
	var v9 []string
	for _, r := range ro9 {
		f, _ := r.Precio.Float64Value()
		v9 = append(v9, fmt.Sprintf("%d (%.2f)", r.ID, f.Float64))
	}
	fmt.Printf("orden=\"precio\" → %v (err %v)\n", v9, err)
	r9, _ := ids(pool, "SELECT id, nombre, precio FROM productos WHERE categoria='cafe' ORDER BY precio, id LIMIT 5")
	fmt.Printf("ordenado de verdad por precio → %v\n", r9)

	titulo("10 · El plan: la consulta de sqlc con todo opcional contra la que arma Squirrel")
	conn, err := pool.Acquire(ctx)
	if err != nil {
		panic(err)
	}
	defer conn.Release()
	s8b, _, _ := conSquirrel(Filtro{PrecioMax: 1.5, PorPagina: 100000})
	for _, modo := range []string{"force_custom_plan", "force_generic_plan", "auto"} {
		conn.Exec(ctx, "SET plan_cache_mode = "+modo)
		fmt.Printf("### plan_cache_mode = %s\n\n", modo)
		for _, l := range []struct{ n, q, exec string }{
			{"sqlc", bd.SqlBuscar, "EXECUTE p('cafe', NULL, NULL, 100000)"},
			{"Squirrel", s8, "EXECUTE p('cafe')"},
			{"sqlc, solo precio <= 1.5", bd.SqlBuscar, "EXECUTE p(NULL, 1.5, NULL, 100000)"},
			{"Squirrel, solo precio <= 1.5", s8b, "EXECUTE p(1.5)"},
		} {
			conn.Exec(ctx, "DEALLOCATE ALL")
			if _, err := conn.Exec(ctx, "PREPARE p AS "+l.q, pgx.QueryExecModeSimpleProtocol); err != nil {
				fmt.Println("  error en PREPARE:", err)
				continue
			}
			if modo == "auto" { // el modo por defecto: tras 5 ejecuciones, Postgres decide si se queda con el plan genérico
				for i := 0; i < 6; i++ {
					if _, err := conn.Exec(ctx, l.exec, pgx.QueryExecModeSimpleProtocol); err != nil {
						fmt.Println("  error:", err)
					}
				}
				fmt.Printf("(tras 6 ejecuciones) ")
			}
			fmt.Printf("%s · EXPLAIN (COSTS OFF) %s\n", l.n, l.exec)
			opciones := "EXPLAIN (COSTS OFF) "
			if modo == "auto" { // sin tiempos (no es un banco): filas descartadas y bloques leídos, que no dependen del reloj
				opciones = "EXPLAIN (ANALYZE, COSTS OFF, TIMING OFF, SUMMARY OFF, BUFFERS) "
				fmt.Println("  " + opciones)
			}
			r, err := conn.Query(ctx, opciones+l.exec, pgx.QueryExecModeSimpleProtocol)
			if err != nil {
				fmt.Println("  error:", err)
				continue
			}
			for r.Next() {
				var t string
				r.Scan(&t)
				fmt.Println("  " + t)
			}
			if err := r.Err(); err != nil {
				fmt.Println("  error:", err)
			}
			r.Close()
		}
		fmt.Println()
	}
}
