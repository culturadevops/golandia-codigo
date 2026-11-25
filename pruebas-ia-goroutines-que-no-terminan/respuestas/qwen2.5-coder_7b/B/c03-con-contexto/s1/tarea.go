package tarea

import (
	"context"
	"sync"
)

func LlamarConContexto(ctx context.Context, f func() (string, error)) (string, error) {
	resultChan := make(chan result)
	errChan := make(chan error)
	var once sync.Once

	go func() {
		res, err := f()
		once.Do(func() {
			if err != nil {
				errChan <- err
			} else {
				resultChan <- result{res, nil}
			}
		})
	}()

	select {
	case <-ctx.Done():
		return "", ctx.Err()
	case r := <-resultChan:
		return r.value, r.err
	case err := <-errChan:
		return "", err
	}
}

type result struct {
	value string
	err   error
}
