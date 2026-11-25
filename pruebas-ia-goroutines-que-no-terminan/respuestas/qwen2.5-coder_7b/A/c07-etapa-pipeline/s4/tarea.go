package tarea

import (
	"context"
	"math"
)

func Cuadrados(ctx context.Context, in <-chan int) <-chan int {
	out := make(chan int)
	go func() {
		defer close(out)
		for {
			select {
			case <-ctx.Done():
				return
			case num, ok := <-in:
				if !ok {
					return
				}
				out <- int(math.Pow(float64(num), 2))
			}
		}
	}()
	return out
}
