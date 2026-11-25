//go:build refactor

// Refactor SIN cambio de comportamiento visible: el pedido se guarda como «pendiente» ANTES de
// cobrar (para que un cobro nunca quede sin su pedido), y después se actualiza a «pagado» o
// «rechazado». Desde fuera: mismo total, mismo estado final, mismos errores.
package fakesmocks

import (
	"context"
	"fmt"

	"curso-testing/tienda"
)

const Version = "refactor"

type Servicio struct {
	Repo   tienda.Repositorio
	Cobros tienda.Cobrador
}

func (s *Servicio) Crear(ctx context.Context, cliente string, items []tienda.Item) (*tienda.Pedido, error) {
	if err := tienda.Validar(items); err != nil {
		return nil, err
	}
	previos, err := s.Repo.PorCliente(ctx, cliente)
	if err != nil {
		return nil, err
	}
	total := tienda.Total(items, len(previos) == 0)
	p := &tienda.Pedido{Cliente: cliente, Items: items, Total: total, Estado: "pendiente"}
	if err := s.Repo.Guardar(ctx, p); err != nil {
		return nil, err
	}
	id, err := s.Cobros.Cobrar(ctx, cliente, total)
	if err != nil {
		p.Estado = "rechazado"
		s.Repo.Guardar(ctx, p)
		return nil, fmt.Errorf("cobrando: %w", err)
	}
	p.Estado, p.CobroID = "pagado", id
	if err := s.Repo.Guardar(ctx, p); err != nil {
		return nil, err
	}
	return p, nil
}
