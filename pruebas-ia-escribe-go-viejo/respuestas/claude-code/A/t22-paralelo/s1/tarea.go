package tarea

import "sync"

// ProcesarTodos llama a f con cada item en su propia goroutine, espera a que
// terminen todas y devuelve los resultados en el mismo orden que items.
func ProcesarTodos(items []string, f func(string) int) []int {
	resultados := make([]int, len(items))
	var wg sync.WaitGroup
	wg.Add(len(items))
	for i, item := range items {
		go func(i int, item string) {
			defer wg.Done()
			resultados[i] = f(item)
		}(i, item)
	}
	wg.Wait()
	return resultados
}
