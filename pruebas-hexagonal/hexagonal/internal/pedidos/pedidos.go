// Package pedidos es el CENTRO del hexágono: el dominio y sus reglas, sin saber nada de HTTP ni de Postgres.
package pedidos

import (
	"context"
	"errors"
)

// Pedido es el dominio. Nada de etiquetas de base de datos ni de JSON aquí.
type Pedido struct {
	ID       int64
	Cliente  string
	Unidades int
	Precio   int // centavos por unidad
}

func (p Pedido) Total() int { return p.Unidades * p.Precio }

var (
	ErrSinCliente   = errors.New("el pedido necesita un cliente")
	ErrSinUnidades  = errors.New("el pedido necesita al menos una unidad")
	ErrDemasiado    = errors.New("el total pasa del límite de 1.000.000 de centavos")
	ErrNoEncontrado = errors.New("pedido no encontrado")
)

// Repositorio es un PUERTO: lo que el dominio necesita del mundo, dicho por el dominio.
// Los adaptadores (Postgres, memoria) lo implementan; el dominio no sabe cuál le tocó.
type Repositorio interface {
	Guardar(ctx context.Context, p Pedido) (int64, error)
	Buscar(ctx context.Context, id int64) (Pedido, error)
}

// Servicio es el caso de uso. Recibe el puerto, no un adaptador concreto.
type Servicio struct{ repo Repositorio }

func NuevoServicio(r Repositorio) *Servicio { return &Servicio{repo: r} }

func (s *Servicio) Crear(ctx context.Context, cliente string, unidades, precio int) (Pedido, error) {
	p := Pedido{Cliente: cliente, Unidades: unidades, Precio: precio}
	switch {
	case p.Cliente == "":
		return Pedido{}, ErrSinCliente
	case p.Unidades < 1:
		return Pedido{}, ErrSinUnidades
	case p.Total() > 1_000_000:
		return Pedido{}, ErrDemasiado
	}
	id, err := s.repo.Guardar(ctx, p)
	if err != nil {
		return Pedido{}, err
	}
	p.ID = id
	return p, nil
}

func (s *Servicio) Buscar(ctx context.Context, id int64) (Pedido, error) { return s.repo.Buscar(ctx, id) }
