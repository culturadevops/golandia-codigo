//go:build ignore

// 2 · Apagado ordenado: la diferencia entre un binario de juguete y uno de produccion.
// Docker y Kubernetes NO matan tu contenedor: le mandan SIGTERM y esperan unos segundos.
// Lo que hagas en esos segundos es lo que decide si pierdes peticiones a medias o no.
package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	// signal.NotifyContext: el contexto se cancela solo cuando llega la senal.
	// Son DOS lineas para lo que en otros lenguajes es un manejador global.
	ctx, parar := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer parar()

	fmt.Printf("servidor arrancado, PID %d — esperando trabajo\n", os.Getpid())

	// Simulamos que alguien nos manda SIGTERM desde fuera a los 300 ms,
	// que es justo lo que hace `docker stop` o el kubelet.
	go func() {
		time.Sleep(300 * time.Millisecond)
		fmt.Println("[el sistema manda SIGTERM]")
		_ = syscall.Kill(os.Getpid(), syscall.SIGTERM)
	}()

	<-ctx.Done() // aqui esperamos: ni bucles ni sleeps
	fmt.Println("recibida la senal: dejo de aceptar peticiones nuevas")

	// Y ahora el trabajo que de verdad importa: terminar lo que estaba a medias.
	fmt.Println("terminando las peticiones en vuelo...")
	time.Sleep(200 * time.Millisecond)
	fmt.Println("cerrada la base de datos")
	fmt.Println("adios limpiamente")
}
