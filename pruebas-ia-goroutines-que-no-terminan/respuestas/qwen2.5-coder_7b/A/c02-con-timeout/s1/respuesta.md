```go
package tarea

import (
	"errors"
	"sync"
	"time"
)

var ErrTimeout = errors.New("timeout")

func ConTimeout(d time.Duration, f func() int) (int, error) {
	resultChan := make(chan int)
	errChan := make(chan error)

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		select {
		case resultChan <- f():
		case <-time.After(d):
			errChan <- ErrTimeout
		}
	}()

	wg.Wait()
	select {
	case result := <-resultChan:
		return result, nil
	case err := <-errChan:
		return 0, err
	}
}
```