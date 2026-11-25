```go
package tarea

import "context"

type resultado struct {
	valor string
	err   error
}

// LlamarConContexto ejecuta f en otra goroutine y devuelve su resultado.
// Si ctx termina antes de que f finalice, devuelve "" y ctx.Err().
func LlamarConContexto(ctx context.Context, f func() (string, error)) (string, error) {
	if err := ctx.Err(); err != nil {
		return "", err
	}

	// Canal con buffer para que la goroutine no quede bloqueada
	// si nadie recibe el resultado tras cancelarse el contexto.
	ch := make(chan resultado, 1)

	go func() {
		v, err := f()
		ch <- resultado{valor: v, err: err}
	}()

	select {
	case r := <-ch:
		return r.valor, r.err
	case <-ctx.Done():
		return "", ctx.Err()
	}
}
```