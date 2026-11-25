package tarea

import (
	"context"
	"sync"
)

// FanOutFanIn reparte los valores de entrada entre `workers` goroutines que
// aplican f y junta todos los resultados en un único canal (sin orden
// garantizado). El canal devuelto se cierra cuando todos los workers terminan.
//
// Si ctx se cancela, el productor y los workers dejan de enviar y salen, y el
// canal de salida se cierra. Si el consumidor deja de leer antes de agotar el
// canal, debe cancelar ctx para que ninguna goroutine quede bloqueada.
func FanOutFanIn(ctx context.Context, entrada []int, workers int, f func(int) int) <-chan int {
	if workers < 1 {
		workers = 1
	}
	if n := len(entrada); n > 0 && workers > n {
		workers = n
	}

	trabajos := make(chan int)
	salida := make(chan int)

	// Productor: reparte la entrada y cierra trabajos al terminar o al cancelar.
	go func() {
		defer close(trabajos)
		for _, v := range entrada {
			select {
			case trabajos <- v:
			case <-ctx.Done():
				return
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
					case salida <- r:
					case <-ctx.Done():
						return
					}
				}
			}
		}()
	}

	// Cierra la salida cuando todos los workers han terminado.
	go func() {
		wg.Wait()
		close(salida)
	}()

	return salida
}
