```go
package tarea

import (
	"context"
	"sync"
	"sync/atomic"
)

// Consumir arranca `consumidores` goroutines que leen de trabajos y llaman a f
// con cada valor, hasta que trabajos se cierre o ctx se cancele. Espera a que
// terminen todas antes de volver. Devuelve ctx.Err() si se canceló, o nil si
// trabajos se cerró.
func Consumir(ctx context.Context, trabajos <-chan int, consumidores int, f func(int)) error {
	if consumidores < 1 {
		consumidores = 1
	}

	var (
		wg        sync.WaitGroup
		cancelado atomic.Bool
	)

	wg.Add(consumidores)
	for i := 0; i < consumidores; i++ {
		go func() {
			defer wg.Done()
			for {
				select {
				case <-ctx.Done():
					cancelado.Store(true)
					return
				case t, ok := <-trabajos:
					if !ok {
						return
					}
					// select elige al azar si ambos están listos; damos
					// prioridad a la cancelación antes de procesar.
					if ctx.Err() != nil {
						cancelado.Store(true)
						return
					}
					f(t)
				}
			}
		}()
	}

	wg.Wait()

	if cancelado.Load() {
		return ctx.Err()
	}
	return nil
}
```