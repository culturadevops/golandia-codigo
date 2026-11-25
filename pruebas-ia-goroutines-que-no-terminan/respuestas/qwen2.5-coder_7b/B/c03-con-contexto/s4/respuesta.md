```go
package tarea

import (
	"context"
	"sync"
)

func LlamarConContexto(ctx context.Context, f func() (string, error)) (string, error) {
	resultChan := make(chan result)
	var wg sync.WaitGroup

	wg.Add(1)
	go func() {
		defer wg.Done()
		result, err := f()
		resultChan <- result
	}()

	select {
	case <-ctx.Done():
		return "", ctx.Err()
	case result := <-resultChan:
		return result, nil
	}

	wg.Wait()
	return "", ctx.Err() // This line should never be reached
}

type result struct {
	value string
	err   error
}
```