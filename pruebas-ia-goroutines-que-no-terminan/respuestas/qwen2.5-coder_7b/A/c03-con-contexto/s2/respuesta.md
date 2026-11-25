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
		res, err := f()
		resultChan <- result{res, err}
	}()

	select {
	case <-ctx.Done():
		return "", ctx.Err()
	case res := <-resultChan:
		return res.res, res.err
	}
}

type result struct {
	res string
	err error
}
```