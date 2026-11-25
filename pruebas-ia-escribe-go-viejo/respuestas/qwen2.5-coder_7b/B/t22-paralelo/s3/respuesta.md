```go
package tarea

import (
	"sync"
)

func ProcesarTodos(items []string, f func(string) int) []int {
	var wg sync.WaitGroup
	results := make([]int, len(items))

	for i, item := range items {
		wg.Add(1)
		go func(i int, item string) {
			defer wg.Done()
			results[i] = f(item)
		}(i, item)
	}

	wg.Wait()
	return results
}
```