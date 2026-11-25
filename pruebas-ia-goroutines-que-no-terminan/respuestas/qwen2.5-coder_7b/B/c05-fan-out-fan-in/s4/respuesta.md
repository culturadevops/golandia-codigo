```go
package tarea

import (
	"context"
	"sync"
)

func FanOutFanIn(ctx context.Context, entrada []int, workers int, f func(int) int) <-chan int {
	output := make(chan int)
	var wg sync.WaitGroup

	// Función para procesar los datos
	processData := func(in <-chan int, out chan<- int, workerID int) {
		defer wg.Done()
		for n := range in {
			select {
			case <-ctx.Done():
				return
			default:
				out <- f(n)
			}
		}
	}

	// Canal intermedio para recoger los resultados de los workers
	bufferedChan := make(chan int, len(entrada))
	for i := 0; i < workers; i++ {
		wg.Add(1)
		go processData(entrada[i*len(entrada)/workers:(i+1)*len(entrada)/workers], bufferedChan, i)
	}

	// Goroutine para juntar los resultados
	go func() {
		wg.Wait()
		close(output)
	}()

	// Goroutine para transferir datos del canal intermedio al canal de salida
	go func() {
		for n := range bufferedChan {
			output <- n
		}
	}()

	return output
}
```