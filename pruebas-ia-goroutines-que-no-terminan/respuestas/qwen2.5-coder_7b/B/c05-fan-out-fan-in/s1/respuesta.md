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

	// Función para procesar los datos
	processData := func(input <-chan int, output chan<- int) {
		defer wg.Done()
		for n := range input {
			select {
			case <-ctx.Done():
				return
			default:
				output <- f(n)
			}
		}
	}

	// Goroutine para enviar datos al procesador
	inputChan := make(chan int)
	go func() {
		defer close(inputChan)
		for _, n := range entrada {
			select {
			case <-ctx.Done():
				return
			case inputChan <- n:
			}
		}
	}()

	// Goroutines para procesar los datos
	for i := 0; i < workers; i++ {
		go processData(inputChan, output)
	}

	// Goroutine para cerrar el canal de salida cuando todas las goroutines de procesamiento terminen
	go func() {
		wg.Wait()
		close(output)
	}()

	return output
}
```