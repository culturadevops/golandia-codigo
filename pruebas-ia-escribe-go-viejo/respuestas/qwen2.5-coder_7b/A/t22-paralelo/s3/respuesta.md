```go
package tarea

import (
	"sync"
)

func ProcesarTodos(items []string, f func(string) int) []int {
	var wg sync.WaitGroup
	results := make([]int, len(items))
	resultChan := make(chan int, len(items))

	for i, item := range items {
		wg.Add(1)
		go func(i int, item string) {
			defer wg.Done()
			resultChan <- f(item)
		}(i, item)
	}

	go func() {
		wg.Wait()
		close(resultChan)
	}()

	for i := range resultChan {
		results[i] = i
	}

	return results
}
```