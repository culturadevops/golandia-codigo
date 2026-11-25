// Lección 5 · rows sin cerrar: el pool que se agota.
//
// La trampa: una función que busca «el primero» hace rows.Next() una vez y retorna. No
// cierra rows, así que la conexión no vuelve al pool. No hay error. Cuando se acaban las
// conexiones, la siguiente consulta se queda esperando turno... para siempre.
//
// Hallazgo preparando la lección: si la consulta lleva un context y ese context se cancela,
// database/sql cierra el rows por ti. La fuga de verdad es db.Query sin context.
package main

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"curso-sql/comun"
)

// primeroMal devuelve el primer servidor de un rol y se olvida de rows.
func primeroMal(db *sql.DB, rol string) (string, error) {
	rows, err := db.Query("SELECT nombre FROM servidores WHERE rol = $1 ORDER BY id", rol)
	if err != nil {
		return "", err
	}
	if rows.Next() {
		var n string
		rows.Scan(&n)
		return n, nil // ⛔ rows sigue abierto: la conexión no vuelve al pool
	}
	return "", nil
}

func primeroBien(db *sql.DB, rol string) (string, error) {
	rows, err := db.Query("SELECT nombre FROM servidores WHERE rol = $1 ORDER BY id", rol)
	if err != nil {
		return "", err
	}
	defer rows.Close() // ✅
	if rows.Next() {
		var n string
		return n, rows.Scan(&n)
	}
	return "", rows.Err()
}

// primeroCtx es el mismo olvido que primeroMal, pero con un context que se cancela al volver.
func primeroCtx(db *sql.DB, rol string) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	rows, err := db.QueryContext(ctx, "SELECT nombre FROM servidores WHERE rol = $1 ORDER BY id", rol)
	if err != nil {
		return "", err
	}
	if rows.Next() {
		var n string
		rows.Scan(&n)
		return n, nil // tampoco cierra rows... pero el cancel() de arriba sí
	}
	return "", nil
}

func ronda(titulo string, f func(*sql.DB, string) (string, error)) {
	comun.Titulo(titulo)
	db := comun.Abrir()
	defer db.Close()
	db.SetMaxOpenConns(3)
	for i := 1; i <= 5; i++ {
		hecho := make(chan string, 1)
		go func() {
			n, err := f(db, "web")
			if err != nil {
				n = "ERROR: " + err.Error()
			}
			hecho <- n
		}()
		select {
		case n := <-hecho:
			fmt.Printf("  llamada %d -> %s   (conexiones en uso: %d de 3)\n", i, n, db.Stats().InUse)
		case <-time.After(2 * time.Second):
			st := db.Stats()
			fmt.Printf("  llamada %d -> sigue esperando tras 2 s   (en uso: %d de 3, esperando turno: %d)\n",
				i, st.InUse, st.WaitCount)
			fmt.Println("  Ni un error. Sin timeout, esta llamada espera para siempre.")
			return
		}
	}
}

func main() {
	db := comun.Abrir()
	comun.Sembrar(db)
	db.Close()

	ronda("5.1 · Sin rows.Close(), pool de 3 conexiones", primeroMal)
	ronda("5.2 · Con defer rows.Close()", primeroBien)
	ronda("5.3 · El mismo olvido, con un context que se cancela", primeroCtx)
}
