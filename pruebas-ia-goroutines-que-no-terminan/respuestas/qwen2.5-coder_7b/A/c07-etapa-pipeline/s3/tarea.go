package tarea

import (
	"context"
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
				out <- num * num
			}
		}
	}()
	return out
}
