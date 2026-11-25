// Package pedidos tiene la lógica de negocio. Va en internal/: nadie fuera de este módulo puede importarlo.
package pedidos

import (
	"errors"
	"fmt"
)

type Pedido struct {
	ID       int
	Producto string
	Cantidad int
}

// Almacen es lo que pedidos necesita para guardar. Lo define quien lo usa, no quien lo implementa.
type Almacen interface {
	Guardar(Pedido) (int, error)
}

type Servicio struct{ A Almacen }

func (s Servicio) Crear(producto string, cantidad int) (Pedido, error) {
	if cantidad <= 0 {
		return Pedido{}, errors.New("la cantidad debe ser mayor que cero")
	}
	p := Pedido{Producto: producto, Cantidad: cantidad}
	id, err := s.A.Guardar(p)
	if err != nil {
		return Pedido{}, fmt.Errorf("guardar pedido: %w", err)
	}
	p.ID = id
	return p, nil
}
