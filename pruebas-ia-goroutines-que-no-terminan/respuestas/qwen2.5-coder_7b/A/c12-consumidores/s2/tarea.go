package tarea

import (
	"context"
	"sync"
)

func Consumir(ctx context.Context, trabajos <-chan int, consumidores int, f func(int)) error {
	var wg sync.WaitGroup
	wg.Add(consumidores)

	for i := 0; i < consumidores; i++ {
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

	wg.Wait()
	if ctx.Err() != nil {
		return ctx.Err()
	}
	return nil
}
