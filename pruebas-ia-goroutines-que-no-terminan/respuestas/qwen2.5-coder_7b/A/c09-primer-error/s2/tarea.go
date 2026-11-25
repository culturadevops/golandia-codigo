package tarea

import (
	"errors"
)

func ProcesarTodos(items []string, f func(string) error) error {
	errChan := make(chan error, len(items))

	for _, item := range items {
		go func(item string) {
			errChan <- f(item)
		}(item)
	}

	for i := 0; i < len(items); i++ {
		if err := <-errChan; err != nil {
			return err
		}
	}

	return nil
}
