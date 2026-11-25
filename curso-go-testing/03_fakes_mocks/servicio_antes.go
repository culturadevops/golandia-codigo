//go:build !refactor && !conbug

// Lección 3 · La versión ORIGINAL de Crear.
//
// Las mismas 12 pruebas (6 con el fake, 6 con mocks de gomock) corren contra tres versiones
// del servicio, elegidas con una etiqueta de compilación:
//
//	go test ./03_fakes_mocks                    -> esta (antes)
//	go test -tags refactor ./03_fakes_mocks     -> refactor que NO cambia el comportamiento
//	go test -tags conbug   ./03_fakes_mocks     -> un fallo de verdad
package fakesmocks

import (
	"context"
	"fmt"

	"curso-testing/tienda"
)

const Version = "antes"

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
	id, err := s.Cobros.Cobrar(ctx, cliente, total)
	if err != nil {
		return nil, fmt.Errorf("cobrando: %w", err)
	}
	p := &tienda.Pedido{Cliente: cliente, Items: items, Total: total, Estado: "pagado", CobroID: id}
	if err := s.Repo.Guardar(ctx, p); err != nil {
		return nil, err
	}
	return p, nil
}
