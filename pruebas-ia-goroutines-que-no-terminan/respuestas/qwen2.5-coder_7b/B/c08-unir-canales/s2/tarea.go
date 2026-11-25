package tarea

import (
	"context"
	"sync"
)

func Unir(ctx context.Context, canales ...<-chan int) <-chan int {
	result := make(chan int)
	closeChan := make(chan struct{})

	go func() {
		defer close(result)
		defer close(closeChan)

		wg := sync.WaitGroup{}
		wg.Add(len(canales))

		for _, ch := range canales {
			go func(ch <-chan int) {
				defer wg.Done()
				for {
					select {
					case <-ctx.Done():
						return
					case v, ok := <-ch:
						if !ok {
							return
						}
						select {
						case <-ctx.Done():
							return
						case result <- v:
						}
					}
				}
			}(ch)
		}

		wg.Wait()
	}()

	go func() {
		<-closeChan
		close(result)
	}()

	return result
}
