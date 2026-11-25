```go
package tarea

import (
	"context"
	"sync"
)

func ProcesarConWorkers(ctx context.Context, trabajos []int, workers int, f func(int) int) ([]int, error) {
	resultados := make([]int, len(trabajos))
	resultChan := make(chan int, len(trabajos))
	errChan := make(chan error, 1)

	var wg sync.WaitGroup
	wg.Add(workers)

	for i := 0; i < workers; i++ {
		go func() {
			defer wg.Done()
			for {
				select {
				case <-ctx.Done():
					return
				case job, ok := <-resultChan:
					if !ok {
						return
					}
					resultados[job] = f(trabajos[job])
				}
			}
		}()
	}

	go func() {
		defer close(resultChan)
		for i, job := range trabajos {
			select {
			case <-ctx.Done():
				return
			case resultChan <- i:
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
```