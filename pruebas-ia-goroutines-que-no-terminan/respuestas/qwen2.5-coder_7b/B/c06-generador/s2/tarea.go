package tarea

import (
	"context"
	"sync"
)

func Generar(ctx context.Context, n int) <-chan int {
	ch := make(chan int)
	var wg sync.WaitGroup

	wg.Add(1)
	go func() {
		defer wg.Done()
		for i := 0; i < n; i++ {
			select {
			case <-ctx.Done():
				return
			case ch <- i:
			}
		}
		close(ch)
	}()

	return ch
}
