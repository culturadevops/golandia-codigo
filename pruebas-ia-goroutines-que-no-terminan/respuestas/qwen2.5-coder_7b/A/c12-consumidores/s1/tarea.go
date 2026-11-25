package tarea

import (
	"context"
	"sync"
)

func Consumir(ctx context.Context, trabajos <-chan int, consumidores int, f func(int)) error {
	var wg sync.WaitGroup
	errChan := make(chan error, consumidores)

	for i := 0; i < consumidores; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for {
				select {
				case <-ctx.Done():
					errChan <- ctx.Err()
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
		wg.Wait()
		close(errChan)
	}()

	select {
	case err := <-errChan:
		return err
	case <-ctx.Done():
		return ctx.Err()
	}
}
