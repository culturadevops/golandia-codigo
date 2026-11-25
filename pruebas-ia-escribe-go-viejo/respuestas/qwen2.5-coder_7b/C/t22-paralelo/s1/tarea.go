package tarea

import (
	"errors"
	"sync"
)

func ProcesarTodos(items []string, f func(string) int) ([]int, error) {
	var wg sync.WaitGroup
	results := make([]int, len(items))
	wg.Add(len(items))

	for i, item := range items {
		i := i
		go func() {
			results[i] = f(item)
			wg.Done()
		}()
	}

	wg.Wait()

	return results, nil
}
