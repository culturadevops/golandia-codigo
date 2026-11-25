package tarea

import (
	"context"
	"sync"
)

func Consumir(ctx context.Context, trabajos <-chan int, consumidores int, f func(int)) error {
	var wg sync.WaitGroup
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	for i := 0; i < consumidores; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for {
				select {
				case <-ctx.Done():
					return
				case trabajo, ok := <-trabajos:
					if !ok {
						return
					}
					f(trabajo)
				}
			}
		}()
	}

	go func() {
		<-ctx.Done()
		close(trabajos)
	}()

	wg.Wait()
	return ctx.Err()
}
