```go
package tarea

import (
	"context"
	"sync"
)

func ProcesarConWorkers(ctx context.Context, trabajos []int, workers int, f func(int) int) ([]int, error) {
	if len(trabajos) == 0 {
		return []int{}, nil
	}

	resultados := make([]int, len(trabajos))
	entradas := make(chan int, len(trabajos))
	salidas := make(chan int, len(trabajos))
	var wg sync.WaitGroup

	// Iniciar workers
	for i := 0; i < workers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for trabajo := range entradas {
				select {
				case <-ctx.Done():
					return
				case salidas <- f(trabajo):
				}
			}
		}()
	}

	// Enviar trabajos a workers
	go func() {
		defer close(entradas)
		for i, trabajo := range trabajos {
			select {
			case <-ctx.Done():
				return
			case entradas <- trabajo:
			}
		}
	}()

	// Recibir resultados de workers
	go func() {
		defer close(salidas)
		for i := range salidas {
			resultados[i] = i
		}
	}()

	// Esperar a que todos los workers terminen
	wg.Wait()

	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
		return resultados, nil
	}
}
```