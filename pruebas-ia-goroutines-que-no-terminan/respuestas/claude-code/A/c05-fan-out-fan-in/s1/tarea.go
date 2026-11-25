package tarea

import (
	"context"
	"sync"
)

// FanOutFanIn reparte los valores de entrada entre `workers` goroutines que
// aplican f y junta todos los resultados en un único canal. El canal devuelto
// se cierra cuando todos los workers terminan. Si ctx se cancela, la
// distribución y los workers se detienen y el canal se cierra.
func FanOutFanIn(ctx context.Context, entrada []int, workers int, f func(int) int) <-chan int {
	if workers < 1 {
		workers = 1
	}

	trabajos := make(chan int)
	salida := make(chan int)

	// Fan-out: productor que envía la entrada a los workers.
	go func() {
		defer close(trabajos)
		for _, v := range entrada {
			select {
			case <-ctx.Done():
				return
			case trabajos <- v:
			}
		}
	}()

	var wg sync.WaitGroup
	wg.Add(workers)
	for i := 0; i < workers; i++ {
		go func() {
			defer wg.Done()
			for {
				select {
				case <-ctx.Done():
					return
				case v, ok := <-trabajos:
					if !ok {
						return
					}
					r := f(v)
					select {
					case <-ctx.Done():
						return
					case salida <- r:
					}
				}
			}
		}()
	}

	// Fan-in: cierra la salida cuando terminan todos los workers.
	go func() {
		wg.Wait()
		close(salida)
	}()

	return salida
}
