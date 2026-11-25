Buffered channel so the goroutine doesn't leak when blocked on send.

```go
package tarea

import (
	"errors"
	"time"
)

// ErrTimeout se devuelve cuando f no termina dentro del plazo indicado.
var ErrTimeout = errors.New("timeout")

// ConTimeout ejecuta f en otra goroutine y devuelve su resultado.
// Si f tarda más que d, devuelve 0 y ErrTimeout sin esperar a f.
func ConTimeout(d time.Duration, f func() int) (int, error) {
	// Canal con buffer para que la goroutine no quede bloqueada
	// (y no haya fuga) si nadie recibe el resultado tras el timeout.
	resultado := make(chan int, 1)

	go func() {
		resultado <- f()
	}()

	timer := time.NewTimer(d)
	defer timer.Stop()

	select {
	case v := <-resultado:
		return v, nil
	case <-timer.C:
		return 0, ErrTimeout
	}
}
```