// Lección 10: una readiness sin plazo, contra una base que no contesta (docker pause).
//
//	APP_BD_URL=... go run ./lecciones/10_salud
//
// Conecta, escribe «conectado» y espera una línea en la entrada: en ese momento
// ejecutar_todo.py ya ha pausado el contenedor de Postgres.
package main

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"time"

	"tareas/internal/bd"
)

func main() {
	ctx := context.Background()
	pool, err := bd.Abrir(ctx, os.Getenv("APP_BD_URL"))
	if err != nil {
		panic(err)
	}
	fmt.Println("conectado")
	_, _ = bufio.NewReader(os.Stdin).ReadString('\n')

	for _, plazo := range []time.Duration{0, time.Second} {
		hecho := make(chan error, 1)
		t0 := time.Now()
		go func() {
			c := ctx
			if plazo > 0 {
				var cancel context.CancelFunc
				c, cancel = context.WithTimeout(ctx, plazo)
				defer cancel()
			}
			hecho <- pool.Ping(c)
		}()
		nombre := "Ping(context.Background())"
		if plazo > 0 {
			nombre = "Ping con plazo de 1 s     "
		}
		select {
		case err := <-hecho:
			fmt.Printf("  %s -> en %d ms: %v\n", nombre, time.Since(t0).Round(10*time.Millisecond).Milliseconds(), err)
		case <-time.After(5 * time.Second):
			fmt.Printf("  %s -> sin respuesta a los 5 s (la sonda se queda colgada)\n", nombre)
		}
	}
	os.Exit(0) // el Ping colgado sigue ahí: no se espera a pool.Close
}
