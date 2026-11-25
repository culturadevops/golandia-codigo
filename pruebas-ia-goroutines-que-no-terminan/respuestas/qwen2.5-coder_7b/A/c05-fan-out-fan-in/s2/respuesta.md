```go
package tarea

import (
	"context"
	"sync"
)

func FanOutFanIn(ctx context.Context, entrada []int, workers int, f func(int) int) <-chan int {
	output := make(chan int)
	var wg sync.WaitGroup

	// Goroutine que distribuye los trabajos entre los workers
	go func() {
		defer close(output)
		chunks := partition(entrada, workers)
		for _, chunk := range chunks {
			wg.Add(1)
			go func(chunk []int) {
				defer wg.Done()
				for _, num := range chunk {
					select {
					case <-ctx.Done():
						return
					case output <- f(num):
					}
				}
			}(chunk)
		}
	}()

	// Espera a que todos los trabajos terminen
	go func() {
		wg.Wait()
		close(output)
	}()

	return output
}

// Función auxiliar para dividir el slice en partes
func partition(slice []int, n int) [][]int {
	chunks := make([][]int, n)
	for i := range chunks {
		chunks[i] = make([]int, 0)
	}
	for i, num := range slice {
		chunks[i%n] = append(chunks[i%n], num)
	}
	return chunks
}
```