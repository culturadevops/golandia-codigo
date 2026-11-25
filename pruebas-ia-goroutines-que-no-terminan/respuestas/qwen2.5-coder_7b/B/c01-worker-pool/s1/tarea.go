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
	resultChan := make(chan int, len(trabajos))
	errChan := make(chan error, 1)
	wg := sync.WaitGroup{}

	for i := 0; i < workers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for {
				select {
				case <-ctx.Done():
					return
				case trabajo, ok := <-trabajosChan:
					if !ok {
						return
					}
					resultChan <- f(trabajo)
				}
			}
		}()
	}

	trabajosChan := make(chan int, workers)
	go func() {
		defer close(trabajosChan)
		for i, trabajo := range trabajos {
			select {
			case <-ctx.Done():
				return
			case trabajosChan <- trabajo:
				resultados[i] = <-resultChan
			}
		}
	}()

	go func() {
		wg.Wait()
		close(resultChan)
	}()

	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case err := <-errChan:
		return nil, err
	}
}
