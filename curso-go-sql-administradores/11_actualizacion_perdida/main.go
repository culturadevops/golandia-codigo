// Lección 11 · La actualización perdida.
//
// La trampa: leer el stock en Go, restarle uno en Go y guardar el resultado. Con una
// petición a la vez, perfecto. Con cincuenta a la vez, dos leen el mismo valor y una de
// las ventas desaparece. Sin un solo error.
package main

import (
	"database/sql"
	"fmt"
	"sync"
	"time"

	"curso-sql/comun"
)

const ventas = 50

func reset(db *sql.DB) {
	comun.Debe(db, `DROP TABLE IF EXISTS productos`)
	comun.Debe(db, `CREATE TABLE productos (id int PRIMARY KEY, stock int NOT NULL)`)
	comun.Debe(db, `INSERT INTO productos VALUES (1, 100)`)
}

func stock(db *sql.DB) int {
	var s int
	db.QueryRow(`SELECT stock FROM productos WHERE id = 1`).Scan(&s)
	return s
}

func venderMal(db *sql.DB) {
	var s int
	db.QueryRow(`SELECT stock FROM productos WHERE id = 1`).Scan(&s)
	time.Sleep(2 * time.Millisecond) // lo que tarda tu código entre leer y escribir
	db.Exec(`UPDATE productos SET stock = $1 WHERE id = 1`, s-1)
}

func venderAtomico(db *sql.DB) {
	db.Exec(`UPDATE productos SET stock = stock - 1 WHERE id = 1 AND stock > 0`)
}

func venderForUpdate(db *sql.DB) {
	tx, _ := db.Begin()
	defer tx.Rollback()
	var s int
	tx.QueryRow(`SELECT stock FROM productos WHERE id = 1 FOR UPDATE`).Scan(&s)
	time.Sleep(2 * time.Millisecond)
	tx.Exec(`UPDATE productos SET stock = $1 WHERE id = 1`, s-1)
	tx.Commit()
}

func correr(db *sql.DB, f func(*sql.DB)) (int, time.Duration) {
	reset(db)
	var wg sync.WaitGroup
	t0 := time.Now()
	for range ventas {
		wg.Add(1)
		go func() { defer wg.Done(); f(db) }()
	}
	wg.Wait()
	return stock(db), time.Since(t0)
}

func main() {
	db := comun.Abrir()
	defer db.Close()
	db.SetMaxOpenConns(20)

	comun.Titulo("11.1 · Leer en Go, restar en Go, guardar: 50 ventas sobre 100")
	for i := 1; i <= 3; i++ {
		s, _ := correr(db, venderMal)
		fmt.Printf("  intento %d: esperaba stock 50, quedo %d  -> %d ventas perdidas\n", i, s, s-50)
	}

	comun.Titulo("11.2 · Que reste la base de datos: SET stock = stock - 1")
	s, d := correr(db, venderAtomico)
	fmt.Printf("  quedo %d   (%d ms)\n", s, d.Milliseconds())

	comun.Titulo("11.3 · Si necesitas leer antes de escribir: SELECT ... FOR UPDATE")
	s, d = correr(db, venderForUpdate)
	fmt.Printf("  quedo %d   (%d ms: las ventas hacen cola en la fila)\n", s, d.Milliseconds())
}
