package tarea

import (
	"context"
	"sync"
)

func FanOutFanIn(ctx context.Context, entrada []int, workers int, f func(int) int) <-chan int {
	out := make(chan int)

	var wg sync.WaitGroup
	wg.Add(workers)

	// Fan-out
	entradaCanales := make([]chan int, workers)
	for i := range entradaCanales {
		entradaCanales[i] = make(chan int)
		go func(i int) {
			defer wg.Done()
			for n := range entradaCanales[i] {
				select {
				case <-ctx.Done():
					return
				case out <- f(n):
				}
			}
		}(i)
	}

	// Distribución de entrada
	go func() {
		defer close(out)
		for i, n := range entrada {
			entradaCanales[i % workers] <- n
		}
		for i := range entradaCanales {
			close(entradaCanales[i])
		}
	}()

	// Espera por todas las goroutines
	go func() {
		wg.Wait()
		close(out)
	}()

	return out
}
