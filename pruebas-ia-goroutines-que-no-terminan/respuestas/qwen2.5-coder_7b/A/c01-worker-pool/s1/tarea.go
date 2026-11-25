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
	wg := sync.WaitGroup{}
	mu := sync.Mutex{}

	sem := make(chan struct{}, workers)

	for i, trabajo := range trabajos {
		wg.Add(1)
		sem <- struct{}{}
		go func(i int, trabajo int) {
			defer wg.Done()
			defer func() { <-sem }()
			select {
			case <-ctx.Done():
				return
			default:
				mu.Lock()
				resultados[i] = f(trabajo)
				mu.Unlock()
			}
		}(i, trabajo)
	}

	wg.Wait()
	return resultados, ctx.Err()
}
