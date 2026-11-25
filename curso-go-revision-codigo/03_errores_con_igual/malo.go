//go:build malo

package pedidos

import (
	"errors"
	"fmt"
)

var ErrNoExiste = errors.New("el pedido no existe")

func buscar(id int) error {
	return fmt.Errorf("buscando el pedido %d: %w", id, ErrNoExiste)
}

// Codigo dice qué código HTTP devuelve GET /pedidos/{id}.
func Codigo(id int) int {
	err := buscar(id)
	if err == ErrNoExiste {
		return 404
	}
	if err != nil {
		return 500
	}
	return 200
}
