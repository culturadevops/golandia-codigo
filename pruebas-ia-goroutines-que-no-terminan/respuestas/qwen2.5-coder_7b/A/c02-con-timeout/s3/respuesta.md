```go
package tarea

import (
	"errors"
	"sync"
	"time"
)

func ConTimeout(d time.Duration, f func() int) (int, error) {
	resultChan := make(chan int, 1)
	errorChan := make(chan error, 1)
	var once sync.Once

	go func() {
		once.Do(func() {
			select {
			case resultChan <- f():
			case <-time.After(d):
				errorChan <- errors.New("timeout")
			}
		})
	}()

	select {
	case res := <-resultChan:
		return res, nil
	case err := <-errorChan:
		return 0, err
	}
}
```