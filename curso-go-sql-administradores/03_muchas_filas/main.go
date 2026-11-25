// Lección 3 · Muchas filas: rows.Next y rows.Err.
//
// La trampa: si la consulta falla A MITAD (una división por cero en la fila 5, la conexión
// que se corta, un timeout), rows.Next() devuelve false como si se hubieran acabado las
// filas. Si no miras rows.Err(), tu informe sale con cuatro filas y dice «OK».
package main

import (
	"database/sql"
	"fmt"

	"curso-sql/comun"
)

// Para cada i, 100/(5-i): en i=5 divide por cero. Postgres manda las filas según las
// calcula, así que las cuatro primeras llegan antes que el error.
const consulta = `SELECT i, 100 / (5 - i) FROM generate_series(1, 8) AS i`

func informeMal(db *sql.DB) {
	rows, err := db.Query(consulta)
	if err != nil {
		fmt.Println("  error al lanzar:", err)
		return
	}
	defer rows.Close()
	n := 0
	for rows.Next() {
		var i, v int
		rows.Scan(&i, &v)
		fmt.Printf("    fila %d -> %d\n", i, v)
		n++
	}
	fmt.Printf("  procesadas %d filas de 8 -> informe OK\n", n)
}

func informeBien(db *sql.DB) error {
	rows, err := db.Query(consulta)
	if err != nil {
		return err
	}
	defer rows.Close()
	n := 0
	for rows.Next() {
		var i, v int
		if err := rows.Scan(&i, &v); err != nil {
			return err
		}
		n++
	}
	if err := rows.Err(); err != nil { // ⭐ la línea que falta en casi todos los tutoriales
		return fmt.Errorf("el informe se corto en la fila %d: %w", n+1, err)
	}
	fmt.Printf("  procesadas %d filas\n", n)
	return nil
}

func main() {
	db := comun.Abrir()
	defer db.Close()

	comun.Titulo("3.1 · El bucle de siempre, sin rows.Err()")
	informeMal(db)

	comun.Titulo("3.2 · El mismo bucle, mirando rows.Err()")
	if err := informeBien(db); err != nil {
		fmt.Println("  ERROR:", err)
	}
}
