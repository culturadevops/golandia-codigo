package tarea

import (
	"context"
	"sync"
)

func Unir(ctx context.Context, canales ...<-chan int) <-chan int {
	out := make(chan int)
	var wg sync.WaitGroup

	for _, c := range canales {
		wg.Add(1)
		go func(c <-chan int) {
			defer wg.Done()
			select {
			case <-ctx.Done():
				close(out)
			case <-c:
				close(out)
				return
			}
			for v := range c {
				select {
				case <-ctx.Done():
					close(out)
				case out <- v:
				}
			}
		}(c)
	}

	go func() {
		wg.Wait()
		close(out)
	}()

	return out
}
