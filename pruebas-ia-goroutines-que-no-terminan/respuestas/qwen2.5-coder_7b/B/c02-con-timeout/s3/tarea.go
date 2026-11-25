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
	resultChan := make(chan int)
	errorChan := make(chan error)

	go func() {
		resultChan <- f()
	}()

	select {
	case result := <-resultChan:
		return result, nil
	case <-time.After(d):
		errorChan <- &timeoutError{}
	case err := <-errorChan:
		return 0, err
	}

	// This line should never be reached because either the resultChan or errorChan will always be selected.
	return 0, nil
}
