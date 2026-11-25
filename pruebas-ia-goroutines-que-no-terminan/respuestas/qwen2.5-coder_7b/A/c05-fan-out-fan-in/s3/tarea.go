package tarea

import (
	"context"
	"sync"
)

func FanOutFanIn(ctx context.Context, entrada []int, workers int, f func(int) int) <-chan int {
	result := make(chan int)

	var wg sync.WaitGroup
	wg.Add(workers)

	output := make(chan int)
	go func() {
		wg.Wait()
		close(output)
	}()

	start := 0
	for i := 0; i < workers; i++ {
		end := start + len(entrada)/workers
		if i == workers-1 {
			end = len(entrada)
		}
		go func(start, end int) {
			defer wg.Done()
			for _, v := range entrada[start:end] {
				select {
				case <-ctx.Done():
					return
				case output <- f(v):
				}
			}
		}(start, end)
		start = end
	}

	go func() {
		for v := range output {
			result <- v
		}
		close(result)
	}()

	return result
}
