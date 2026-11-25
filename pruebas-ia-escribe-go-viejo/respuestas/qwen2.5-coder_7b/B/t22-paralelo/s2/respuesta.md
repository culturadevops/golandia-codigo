```go
package tarea

import (
	"sync"
)

func ProcesarTodos(items []string, f func(string) int) []int {
	var wg sync.WaitGroup
	results := make([]int, len(items))
	resultsChan := make(chan int, len(items))

	for i, item := range items {
		wg.Add(1)
		go func(i int, item string) {
			defer wg.Done()
			result := f(item)
			resultsChan <- result
		}(i, item)
	}

	go func() {
		wg.Wait()
		close(resultsChan)
	}()

	for i := range resultsChan {
		results[i] = i
	}

	return results
}
```