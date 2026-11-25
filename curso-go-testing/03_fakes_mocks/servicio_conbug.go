//go:build conbug

// Un fallo DE VERDAD: guarda el pedido como «pendiente», cobra, lo marca «pagado»... y se
// olvida de volver a guardarlo. El cliente pagó y en la base de datos el pedido sigue pendiente.
package fakesmocks

import (
	"context"
	"fmt"

	"curso-testing/tienda"
)

const Version = "conbug"

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
		return nil, fmt.Errorf("cobrando: %w", err)
	}
	p.Estado, p.CobroID = "pagado", id
	return p, nil // falta: s.Repo.Guardar(ctx, p)
}
