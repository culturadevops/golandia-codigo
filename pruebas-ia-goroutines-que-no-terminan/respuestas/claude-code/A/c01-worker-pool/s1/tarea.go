package tarea

import (
	"context"
	"sync"
)

// ProcesarConWorkers aplica f a cada trabajo usando un pool de workers goroutines
// y devuelve los resultados en el mismo orden que trabajos. Si ctx se cancela
// antes de terminar, devuelve ctx.Err() sin esperar a los trabajos pendientes.
func ProcesarConWorkers(ctx context.Context, trabajos []int, workers int, f func(int) int) ([]int, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	if workers < 1 {
		workers = 1
	}
	if workers > len(trabajos) {
		workers = len(trabajos)
	}

	resultados := make([]int, len(trabajos))
	if len(trabajos) == 0 {
		return resultados, nil
	}

	indices := make(chan int)
	var wg sync.WaitGroup

	wg.Add(workers)
	for w := 0; w < workers; w++ {
		go func() {
			defer wg.Done()
			for i := range indices {
				if ctx.Err() != nil {
					return
				}
				// Cada índice lo escribe un único worker, así que no hay carrera.
				resultados[i] = f(trabajos[i])
			}
		}()
	}

	go func() {
		defer close(indices)
		for i := range trabajos {
			select {
			case indices <- i:
			case <-ctx.Done():
				return
			}
		}
	}()

	terminado := make(chan struct{})
	go func() {
		wg.Wait()
		close(terminado)
	}()

	select {
	case <-terminado:
		if err := ctx.Err(); err != nil {
			// Algún worker pudo abandonar trabajos por la cancelación.
			return nil, err
		}
		return resultados, nil
	case <-ctx.Done():
		return nil, ctx.Err()
	}
}
