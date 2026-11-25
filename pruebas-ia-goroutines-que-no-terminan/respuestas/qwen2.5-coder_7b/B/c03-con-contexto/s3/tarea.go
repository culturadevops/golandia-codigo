package tarea

import (
	"context"
	"sync"
)

func LlamarConContexto(ctx context.Context, f func() (string, error)) (string, error) {
	resultChan := make(chan result, 1)
	done := make(chan struct{}, 1)

	go func() {
		result, err := f()
		resultChan <- result{value: result, err: err}
		done <- struct{}{}
	}()

	select {
	case <-ctx.Done():
		return "", ctx.Err()
	case r := <-resultChan:
		<-done
		return r.value, r.err
	}
}

type result struct {
	value string
	err   error
}
