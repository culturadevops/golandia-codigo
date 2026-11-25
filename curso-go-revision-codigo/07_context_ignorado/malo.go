//go:build malo

package pedidos

import (
	"context"
	"time"
)

// esperarStock simula la consulta al almacén: tarda 2 s o hasta que cancelen.
func esperarStock(ctx context.Context) error {
	select {
	case <-time.After(2 * time.Second):
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

// Reservar la llama el handler con el context de la petición.
func Reservar(ctx context.Context) error {
	c, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	return esperarStock(c)
}
