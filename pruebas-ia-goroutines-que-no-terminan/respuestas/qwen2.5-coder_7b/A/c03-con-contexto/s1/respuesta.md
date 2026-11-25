```go
package tarea

import (
	"context"
	"sync"
)

func LlamarConContexto(ctx context.Context, f func() (string, error)) (string, error) {
	var result string
	var err error
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		result, err = f()
	}()
	select {
	case <-ctx.Done():
		return "", ctx.Err()
	case <-wg.Wait():
		return result, err
	}
}
```