package tarea

import (
	"errors"
	"sync"
	"time"
)

func ConTimeout(d time.Duration, f func() int) (int, error) {
	var result int
	var err error
	var wg sync.WaitGroup

	wg.Add(1)
	go func() {
		defer wg.Done()
		result, err = f()
	}()

	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()

	select {
	case <-time.After(d):
		return 0, errors.New("timeout")
	case <-done:
		return result, err
	}
}
