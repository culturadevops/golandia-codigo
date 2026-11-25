package tarea

import (
	"errors"
	"sync"
)

func ProcesarTodos(items []string, f func(string) error) error {
	var wg sync.WaitGroup
	errChan := make(chan error, 1)

	for _, item := range items {
		wg.Add(1)
		go func(item string) {
			defer wg.Done()
			err := f(item)
			if err != nil {
				select {
				case errChan <- err:
				default:
				}
			}
		}(item)
	}

	go func() {
		wg.Wait()
		close(errChan)
	}()

	select {
	case err := <-errChan:
		return err
	case <-errChan:
		return nil
	}
}
