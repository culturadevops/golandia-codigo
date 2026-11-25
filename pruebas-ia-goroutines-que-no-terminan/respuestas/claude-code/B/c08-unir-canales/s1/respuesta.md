```go
package tarea

import (
	"context"
	"sync"
)

// Unir junta todos los valores de los canales de entrada en un solo canal de
// salida. El canal de salida se cierra cuando todos los canales de entrada se
// han cerrado o cuando ctx se cancela. Ninguna goroutine queda bloqueada tras
// la cancelación, aunque nadie siga leyendo del canal de salida.
func Unir(ctx context.Context, canales ...<-chan int) <-chan int {
	salida := make(chan int)
	var wg sync.WaitGroup

	reenviar := func(c <-chan int) {
		defer wg.Done()
		for {
			select {
			case <-ctx.Done():
				return
			case v, ok := <-c:
				if !ok {
					return
				}
				select {
				case salida <- v:
				case <-ctx.Done():
					return
				}
			}
		}
	}

	for _, c := range canales {
		if c == nil {
			continue
		}
		wg.Add(1)
		go reenviar(c)
	}

	go func() {
		wg.Wait()
		close(salida)
	}()

	return salida
}
```