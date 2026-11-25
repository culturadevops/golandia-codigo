```go
package tarea

import (
	"sync"
)

func ProcesarTodos(items []string, f func(string) int) []int {
	var wg sync.WaitGroup
	results := make([]int, len(items))
	ch := make(chan int, len(items))

	for i, item := range items {
		wg.Add(1)
		go func(i int, item string) {
			defer wg.Done()
			result := f(item)
			ch <- result
			results[i] = <-ch
		}(i, item)
	}

	wg.Wait()
	close(ch)

	return results
}
```