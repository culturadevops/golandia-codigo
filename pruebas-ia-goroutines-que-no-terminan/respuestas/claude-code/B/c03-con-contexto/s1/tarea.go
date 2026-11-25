package tarea

import "context"

type resultado struct {
	valor string
	err   error
}

// LlamarConContexto ejecuta f en otra goroutine y devuelve su resultado.
// Si ctx termina antes, devuelve "" y ctx.Err(). El canal tiene buffer 1,
// así que la goroutine nunca se bloquea al enviar aunque nadie lea:
// termina en cuanto f retorna.
func LlamarConContexto(ctx context.Context, f func() (string, error)) (string, error) {
	if err := ctx.Err(); err != nil {
		return "", err
	}

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
