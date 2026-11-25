// Lección 13 · Carga masiva: fila a fila, en una transacción, y con COPY.
//
// La trampa: el bucle de INSERT que en pruebas, con cien filas, va de sobra. Cada INSERT
// suelto es su propia transacción: un viaje de ida y vuelta y una escritura a disco
// confirmada. Con diez mil filas se nota. Con diez millones es una noche.
//
// Cada forma se mide tres veces y se da la mediana: una sola medida de milisegundos miente.
// Y se comparan FILAS POR SEGUNDO: a COPY se le dan 500.000 filas, porque con 10.000 tarda
// unos 5 ms y a esa escala un milisegundo arriba o abajo cambia la cifra un 20 %.
package main

import (
	"context"
	"database/sql"
	"fmt"
	"slices"
	"time"

	"curso-sql/comun"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/stdlib"
)

const ins = `INSERT INTO metricas VALUES ($1, $2, $3)`

func tabla(db *sql.DB) {
	comun.Debe(db, `DROP TABLE IF EXISTS metricas`)
	comun.Debe(db, `CREATE TABLE metricas (servidor text, minuto int, cpu int)`)
}

func contar(db *sql.DB) int {
	var n int
	db.QueryRow(`SELECT count(*) FROM metricas`).Scan(&n)
	return n
}

func filaAFila(db *sql.DB, filas int) {
	for i := range filas {
		db.Exec(ins, fmt.Sprintf("web-%02d", i%20), i, i%100)
	}
}

func enTransaccion(db *sql.DB, filas int) {
	tx, _ := db.Begin()
	st, _ := tx.Prepare(ins)
	for i := range filas {
		st.Exec(fmt.Sprintf("web-%02d", i%20), i, i%100)
	}
	st.Close()
	tx.Commit()
}

func conCopy(db *sql.DB, filas int) {
	conn, _ := db.Conn(context.Background())
	defer conn.Close()
	err := conn.Raw(func(c any) error {
		pc := c.(*stdlib.Conn).Conn() // el *pgx.Conn que hay debajo de database/sql
		_, err := pc.CopyFrom(context.Background(), pgx.Identifier{"metricas"},
			[]string{"servidor", "minuto", "cpu"},
			pgx.CopyFromSlice(filas, func(i int) ([]any, error) {
				return []any{fmt.Sprintf("web-%02d", i%20), i, i % 100}, nil
			}))
		return err
	})
	if err != nil {
		fmt.Println("  COPY:", err)
	}
}

func medir(etiqueta string, db *sql.DB, filas int, f func(*sql.DB, int)) float64 {
	var ts []time.Duration
	var n int
	for range 3 {
		tabla(db)
		t0 := time.Now()
		f(db, filas)
		ts = append(ts, time.Since(t0))
		n = contar(db)
	}
	slices.Sort(ts)
	d := ts[1]
	ritmo := float64(filas) / d.Seconds()
	fmt.Printf("  %-26s %7d filas en %5d ms   %9.0f filas/s   (en la tabla: %d)\n",
		etiqueta, filas, d.Milliseconds(), ritmo, n)
	return ritmo
}

func main() {
	db := comun.Abrir()
	defer db.Close()

	comun.Titulo("13.1 · Tres formas de cargar filas (mediana de 3)")
	lento := medir("fila a fila (autocommit)", db, 10_000, filaAFila)
	medio := medir("en UNA transaccion", db, 10_000, enTransaccion)
	rapido := medir("COPY (pgx.CopyFrom)", db, 500_000, conCopy)
	fmt.Printf("\n  una transaccion frente a fila a fila: %.0f veces mas filas por segundo\n", medio/lento)
	fmt.Printf("  COPY frente a fila a fila:            %.0f veces mas filas por segundo\n", rapido/lento)
}
