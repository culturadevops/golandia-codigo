package tarea

import (
	"context"
	"sync"
)

func ProcesarConWorkers(ctx context.Context, trabajos []int, workers int, f func(int) int) ([]int, error) {
	resultados := make([]int, len(trabajos))
	sem := make(chan struct{}, workers)
	var wg sync.WaitGroup

	for i, trabajo := range trabajos {
		i := i
		sem <- struct{}{} // Señal para bloquear
		wg.Add(1)
		go func() {
			select {
			case <-ctx.Done():
				<-sem // Liberar el semáforo
				wg.Done()
			default:
				resultados[i] = f(trabajo)
				<-sem // Liberar el semáforo
				wg.Done()
			}
		}()
	}

	wg.Wait()
	return resultados, ctx.Err()
}
