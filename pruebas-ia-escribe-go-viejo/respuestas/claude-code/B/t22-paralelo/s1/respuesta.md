```go
package tarea

import "sync"

// ProcesarTodos llama a f con cada item en su propia goroutine y devuelve
// los resultados en el mismo orden que items.
func ProcesarTodos(items []string, f func(string) int) []int {
	resultados := make([]int, len(items))
	var wg sync.WaitGroup
	for i, item := range items {
		wg.Go(func() {
			resultados[i] = f(item)
		})
	}
	wg.Wait()
	return resultados
}
```