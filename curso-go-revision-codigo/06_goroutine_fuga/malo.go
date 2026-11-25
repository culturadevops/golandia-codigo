//go:build malo

package pedidos

import (
	"errors"
	"time"
)

// Precio le pregunta el precio al proveedor, con un plazo.
func Precio(consultar func() int, plazo time.Duration) (int, error) {
	ch := make(chan int)
	go func() { ch <- consultar() }()
	select {
	case p := <-ch:
		return p, nil
	case <-time.After(plazo):
		return 0, errors.New("el proveedor no contestó a tiempo")
	}
}
