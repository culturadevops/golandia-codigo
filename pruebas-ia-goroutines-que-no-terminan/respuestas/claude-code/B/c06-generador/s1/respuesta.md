```go
package tarea

import "context"

// Generar devuelve un canal por el que envía los números de 0 a n-1 y luego
// lo cierra. Si ctx se cancela, deja de enviar y cierra el canal, de modo que
// la goroutine interna nunca queda bloqueada aunque nadie lea del canal.
func Generar(ctx context.Context, n int) <-chan int {
	out := make(chan int)
	go func() {
		defer close(out)
		for i := 0; i < n; i++ {
			select {
			case <-ctx.Done():
				return
			default:
			}
			select {
			case out <- i:
			case <-ctx.Done():
				return
			}
		}
	}()
	return out
}
```