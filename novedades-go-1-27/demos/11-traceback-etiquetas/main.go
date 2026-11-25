// Go 1.27: con go 1.27 en go.mod, el traceback lleva las etiquetas de pprof de cada goroutine.
// Se corre con GOTRACEBACK=all para ver todas las goroutines en el volcado.
package main

import (
	"context"
	"runtime/pprof"
	"time"
)

func atender(ruta, cliente string, bloqueo chan struct{}) {
	pprof.Do(context.Background(), pprof.Labels("ruta", ruta, "cliente", cliente), func(ctx context.Context) {
		go func() { <-bloqueo }() // la goroutine hereda las etiquetas
	})
}

func main() {
	bloqueo := make(chan struct{})
	atender("/api/pedidos", "acme", bloqueo)
	atender("/api/pagos", "globex", bloqueo)
	time.Sleep(10 * time.Millisecond)
	panic("algo salió mal")
}
