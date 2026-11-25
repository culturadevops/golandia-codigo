// Lección 9 · Transacciones.
//
// La trampa: dentro de la transacción, UNA de las sentencias va por db en vez de por tx.
// Compila igual (las dos tienen Exec). Esa sentencia no está en la transacción: el Rollback
// no la deshace, y el inventario deja de cuadrar.
package main

import (
	"database/sql"
	"errors"
	"fmt"
	"time"

	"curso-sql/comun"
)

func preparar(db *sql.DB) {
	comun.Debe(db, `DROP TABLE IF EXISTS almacenes`)
	comun.Debe(db, `CREATE TABLE almacenes (nombre text PRIMARY KEY, stock int NOT NULL)`)
	comun.Debe(db, `INSERT INTO almacenes VALUES ('norte', 100), ('sur', 100)`)
}

func estado(db *sql.DB) string {
	var n, s int
	db.QueryRow(`SELECT (SELECT stock FROM almacenes WHERE nombre='norte'),
	                    (SELECT stock FROM almacenes WHERE nombre='sur')`).Scan(&n, &s)
	return fmt.Sprintf("norte=%d sur=%d total=%d", n, s, n+s)
}

var errDestinoLleno = errors.New("el destino no tiene sitio")

func trasladarMal(db *sql.DB, n int) error {
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	tx.Exec(`UPDATE almacenes SET stock = stock - $1 WHERE nombre = 'norte'`, n)
	db.Exec(`UPDATE almacenes SET stock = stock + $1 WHERE nombre = 'sur'`, n) // ⛔ db, no tx
	// ... una comprobación posterior falla:
	tx.Rollback()
	return errDestinoLleno
}

func trasladarBien(db *sql.DB, n int) error {
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback() // si ya hubo Commit, no hace nada
	if _, err := tx.Exec(`UPDATE almacenes SET stock = stock - $1 WHERE nombre = 'norte'`, n); err != nil {
		return err
	}
	if _, err := tx.Exec(`UPDATE almacenes SET stock = stock + $1 WHERE nombre = 'sur'`, n); err != nil {
		return err
	}
	return errDestinoLleno // la misma comprobación que falla
}

func main() {
	db := comun.Abrir()
	defer db.Close()

	comun.Titulo("9.1 · Una sentencia por db dentro de la transaccion")
	preparar(db)
	fmt.Println("  antes:              ", estado(db))
	fmt.Println("  trasladar 30 ->     ", trasladarMal(db, 30), "-> Rollback")
	fmt.Println("  despues:            ", estado(db))

	comun.Titulo("9.2 · Todo por tx, con defer tx.Rollback()")
	preparar(db)
	fmt.Println("  antes:              ", estado(db))
	fmt.Println("  trasladar 30 ->     ", trasladarBien(db, 30), "-> Rollback")
	fmt.Println("  despues:            ", estado(db))

	comun.Titulo("9.3 · La red de la leccion 8: comprobar ANTES del Commit")
	comun.Sembrar(db)
	tx, _ := db.Begin()
	res, _ := tx.Exec(`UPDATE servidores SET cpu = 0 WHERE nombre LIKE '%' || $1 || '%'`, "")
	n, _ := res.RowsAffected()
	if n != 1 {
		tx.Rollback()
		fmt.Printf("  cambiaria %d filas y esperaba 1 -> Rollback\n", n)
	} else {
		tx.Commit()
	}
	var ceros int
	db.QueryRow(`SELECT count(*) FROM servidores WHERE cpu = 0`).Scan(&ceros)
	fmt.Printf("  servidores con cpu = 0 en la tabla: %d\n", ceros)

	comun.Titulo("9.4 · Y con un pool de UNA conexion, usar db dentro de tx se cuelga")
	preparar(db)
	db.SetMaxOpenConns(1)
	hecho := make(chan error, 1)
	go func() { hecho <- trasladarMal(db, 30) }()
	select {
	case err := <-hecho:
		fmt.Println("  termino:", err)
	case <-time.After(2 * time.Second):
		fmt.Println("  sigue colgado tras 2 s: la unica conexion la tiene la transaccion,")
		fmt.Println("  y db.Exec espera a que quede una libre. Sin error.")
	}
}
