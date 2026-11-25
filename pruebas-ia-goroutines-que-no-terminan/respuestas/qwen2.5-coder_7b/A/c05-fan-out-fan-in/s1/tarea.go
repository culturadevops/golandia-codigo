package tarea

import (
	"context"
	"sync"
)

func FanOutFanIn(ctx context.Context, entrada []int, workers int, f func(int) int) <-chan int {
	output := make(chan int)
	var wg sync.WaitGroup
	wg.Add(workers)

	// Fan out
	worker := func(in <-chan int) {
		defer wg.Done()
		for n := range in {
			select {
			case <-ctx.Done():
				return
			case output <- f(n):
			}
		}
	}

	ch := make(chan int)
	go func() {
		defer close(ch)
		for _, n := range entrada {
			select {
			case <-ctx.Done():
				return
			case ch <- n:
			}
		}
	}()

	for i := 0; i < workers; i++ {
		go worker(ch)
	}

	go func() {
		wg.Wait()
		close(output)
	}()

	return output
}
