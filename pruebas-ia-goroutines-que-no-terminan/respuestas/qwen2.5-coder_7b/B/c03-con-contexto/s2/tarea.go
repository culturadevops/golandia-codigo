package tarea

import (
	"context"
	"sync"
)

func LlamarConContexto(ctx context.Context, f func() (string, error)) (string, error) {
	var result string
	var err error
	var once sync.Once
	done := make(chan struct{})

	go func() {
		result, err = f()
		close(done)
	}()

	select {
	case <-ctx.Done():
		once.Do(func() {
			err = ctx.Err()
		})
	case <-done:
	}

	return result, err
}
