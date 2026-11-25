package tarea

import (
	"context"
	"sync"
)

func ProcesarConWorkers(ctx context.Context, trabajos []int, workers int, f func(int) int) ([]int, error) {
	resultados := make([]int, len(trabajos))
	sem := make(chan struct{}, workers)
	var wg sync.WaitGroup
	errChan := make(chan error, 1)

	for i, trabajo := range trabajos {
		i := i
		sem <- struct{}{}
		wg.Add(1)
		go func() {
			defer func() { <-sem }()
			defer wg.Done()
			select {
			case <-ctx.Done():
				errChan <- ctx.Err()
			default:
				resultados[i] = f(trabajo)
			}
		}()
	}

	go func() {
		wg.Wait()
		close(errChan)
	}()

	if err := <-errChan; err != nil {
		return nil, err
	}

	return resultados, nil
}
