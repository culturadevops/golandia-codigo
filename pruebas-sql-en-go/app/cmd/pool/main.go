// pool: la trampa de rows.Close(). Con un pool de 4 conexiones, salir del bucle sin cerrar las filas
// deja cada conexión ocupada para siempre.
//
//	modos: olvido (db.Query, lee la 1.ª fila y sale sin Close)
//	       todo   (sin Close, pero lee hasta el final: rows.Next()==false cierra solo)
//	       ctx    (sin Close, con QueryContext y cancel() al salir: cancelar el contexto cierra las filas)
//	       bien   (defer rows.Close())
package main

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"os"
	"time"

	"tienda/conexion"
)

const q = "SELECT nombre FROM productos WHERE categoria = $1 ORDER BY id"

func main() {
	modo := os.Args[1]
	db, err := conexion.Abrir(context.Background())
	if err != nil {
		fmt.Fprintln(os.Stderr, "error al arrancar:", err)
		os.Exit(1)
	}
	defer db.Close()
	db.SetMaxOpenConns(4)

	fmt.Printf("modo %s · pool de 4 conexiones\n", modo)
	for i := 1; i <= 6; i++ {
		t := time.Now()
		hecho := make(chan string, 1)
		go func() {
			n, err := primero(db, modo)
			if err != nil {
				n = "ERROR " + err.Error()
			}
			hecho <- n
		}()
		select {
		case n := <-hecho:
			s := db.Stats()
			fmt.Printf("  consulta %d  %.3fs  en uso=%d  %s\n", i, time.Since(t).Seconds(), s.InUse, n)
		case <-time.After(5 * time.Second):
			s := db.Stats()
			fmt.Printf("  consulta %d  sigue esperando a los 5 s  en uso=%d  esperando=%d\n", i, s.InUse, s.WaitCount)
			os.Exit(3)
		}
	}
}

// primero devuelve el primer producto de la categoría.
func primero(db *sql.DB, modo string) (string, error) {
	var rows *sql.Rows
	var err error
	switch modo {
	case "ctx":
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()
		rows, err = db.QueryContext(ctx, q, "cafe")
	default:
		rows, err = db.Query(q, "cafe")
	}
	if err != nil {
		return "", err
	}
	switch modo {
	case "bien":
		defer rows.Close()
	case "todo":
		var p string
		for rows.Next() {
			var n string
			rows.Scan(&n)
			if p == "" {
				p = n
			}
		}
		return p, rows.Err()
	}
	if !rows.Next() { // me basta la primera fila y salgo
		return "", errors.New("sin filas")
	}
	var n string
	err = rows.Scan(&n)
	return n, err
}
