```go
package tarea

import (
	"errors"
	"time"
)

func ConTimeout(d time.Duration, f func() int) (int, error) {
	resultChan := make(chan int)
	errorChan := make(chan error)

	go func() {
		select {
		case resultChan <- f():
		case <-time.After(d):
			errorChan <- errors.New("timeout")
		}
	}()

	select {
	case result := <-resultChan:
		return result, nil
	case err := <-errorChan:
		return 0, err
	}
}
```