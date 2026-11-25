package tarea

import (
	"errors"
	"time"
)

type timeoutError struct{}

func (e *timeoutError) Error() string {
	return "timeout"
}

func ConTimeout(d time.Duration, f func() int) (int, error) {
	resultChan := make(chan int, 1)
	errorChan := make(chan error, 1)

	go func() {
		select {
		case resultChan <- f():
		case <-time.After(d):
			errorChan <- &timeoutError{}
		}
	}()

	select {
	case result := <-resultChan:
		return result, nil
	case err := <-errorChan:
		return 0, err
	}
}
