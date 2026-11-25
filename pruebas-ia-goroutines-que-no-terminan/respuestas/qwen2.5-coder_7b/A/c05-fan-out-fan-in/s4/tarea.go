package tarea

import (
	"context"
	"sync"
)

func FanOutFanIn(ctx context.Context, entrada []int, workers int, f func(int) int) <-chan int {
	output := make(chan int)
	var wg sync.WaitGroup

	// Goroutine para repartir trabajo
	go func() {
		defer close(output)
		ch := make(chan int, workers)
		defer close(ch)

		for _, v := range entrada {
			ch <- v
		}

		for i := 0; i < workers; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				for v := range ch {
					select {
					case <-ctx.Done():
						return
					case output <- f(v):
					}
				}
			}()
		}

		wg.Wait()
	}()

	return output
}
