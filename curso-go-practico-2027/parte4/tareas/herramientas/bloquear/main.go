// bloquear simula una base de datos lenta: bloquea la tabla tareas
// dentro de una transacción y la suelta a los N segundos.
//
//	go run ./herramientas/bloquear 5
package main

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/jackc/pgx/v5"
)

func main() {
	segundos, _ := strconv.Atoi(os.Args[1])
	ctx := context.Background()
	conn, err := pgx.Connect(ctx, os.Getenv("DATABASE_URL"))
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	tx, err := conn.Begin(ctx)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	if _, err := tx.Exec(ctx, `LOCK TABLE tareas IN ACCESS EXCLUSIVE MODE`); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	fmt.Printf("bloquear: tabla tareas bloqueada %d s\n", segundos)
	time.Sleep(time.Duration(segundos) * time.Second)
	tx.Rollback(ctx)
	fmt.Println("bloquear: tabla tareas liberada")
}
