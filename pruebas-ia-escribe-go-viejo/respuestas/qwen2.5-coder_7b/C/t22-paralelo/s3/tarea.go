package tarea

import (
	"errors"
	"sync"
)

func ProcesarTodos(items []string, f func(string) int) ([]int, error) {
	var wg sync.WaitGroup
	results := make([]int, len(items))
	errChan := make(chan error, 1)

	for i := range items {
		wg.Add(1)
		go func(i int, item string) {
			defer wg.Done()
			results[i], err := f(item)
			if err != nil {
				select {
				case errChan <- err:
				default:
				}
			}
		}(i, items[i])
	}

	wg.Wait()
	close(errChan)

	if err := <-errChan; err != nil {
		return nil, err
	}

	return results, nil
}
