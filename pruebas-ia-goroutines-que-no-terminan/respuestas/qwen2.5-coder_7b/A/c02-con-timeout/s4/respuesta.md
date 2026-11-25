```go
package tarea

import (
	"errors"
	"sync"
	"time"
)

func ConTimeout(d time.Duration, f func() int) (int, error) {
	resultChan := make(chan int, 1)
	errChan := make(chan error, 1)
	done := make(chan struct{})

	go func() {
		select {
		case resultChan <- f():
		case <-done:
		}
	}()

	select {
	case res := <-resultChan:
		return res, nil
	case <-time.After(d):
		close(done)
		return 0, errors.New("timeout")
	}
}
```