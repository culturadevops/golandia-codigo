package tarea

import "context"

// Cuadrados lee números de in y envía su cuadrado por el canal devuelto.
// El canal de salida se cierra cuando in se cierra o cuando ctx se cancela.
func Cuadrados(ctx context.Context, in <-chan int) <-chan int {
	out := make(chan int)
	go func() {
		defer close(out)
		for {
			select {
			case <-ctx.Done():
				return
			case n, ok := <-in:
				if !ok {
					return
				}
				select {
				case out <- n * n:
				case <-ctx.Done():
					return
				}
			}
		}
	}()
	return out
}
