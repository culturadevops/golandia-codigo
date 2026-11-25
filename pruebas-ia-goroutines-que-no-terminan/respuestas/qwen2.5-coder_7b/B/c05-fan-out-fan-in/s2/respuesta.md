```go
package tarea

import (
	"context"
	"sync"
)

func FanOutFanIn(ctx context.Context, entrada []int, workers int, f func(int) int) <-chan int {
	output := make(chan int)
	var wg sync.WaitGroup
	wg.Add(workers)

	input := make(chan int)
	go func() {
		for _, num := range entrada {
			input <- num
		}
		close(input)
	}()

	for i := 0; i < workers; i++ {
		go func() {
			defer wg.Done()
			for num := range input {
				select {
				case <-ctx.Done():
					return
				case output <- f(num):
				}
			}
		}()
	}

	go func() {
		wg.Wait()
		close(output)
	}()

	return output
}
```