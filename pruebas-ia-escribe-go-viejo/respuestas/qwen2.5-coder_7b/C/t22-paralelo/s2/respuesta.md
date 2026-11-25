```go
package tarea

import (
	"errors"
	"sync"
)

func ProcesarTodos(items []string, f func(string) int) ([]int, error) {
	if len(items) == 0 {
		return nil, nil
	}

	results := make([]int, len(items))
	var wg sync.WaitGroup
	for i, item := range items {
		i := i
		wg.Add(1)
		go func() {
			results[i] = f(item)
			wg.Done()
		}()
	}
	wg.Wait()

	return results, nil
}
```