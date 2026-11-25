```go
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

	for range items {
		if err := <-errChan; err != nil {
			return err
		}
	}

	return nil
}
```