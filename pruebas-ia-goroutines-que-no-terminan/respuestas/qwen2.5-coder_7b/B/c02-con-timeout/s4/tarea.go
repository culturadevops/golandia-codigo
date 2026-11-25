package tarea

import (
	"errors"
	"time"
)

func ConTimeout(d time.Duration, f func() int) (int, error) {
	resultChan := make(chan int, 1)
	errorChan := make(chan error, 1)

	go func() {
		resultChan <- f()
	}()

	select {
	case result := <-resultChan:
		return result, nil
	case <-time.After(d):
		return 0, errors.New("timeout")
	}
}
