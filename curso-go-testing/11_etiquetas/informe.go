// Lección 11 · Build tags para separar los tests de integración.
//
//	go test ./...                         -> solo los rápidos (sin Docker)
//	go test -tags integracion ./...       -> también los de Postgres en contenedor
//	go vet  -tags integracion ./...       -> que los de integración al menos COMPILEN
//
// Un archivo con //go:build integracion no se compila sin la etiqueta. Si renombras una
// función y no corres con la etiqueta, ese test se queda roto y nadie se entera.
package etiquetas

import (
	"context"

	"curso-testing/tienda"
)

// PagadosDe cuenta los pedidos pagados de un cliente.
func PagadosDe(ctx context.Context, r tienda.Repositorio, cliente string) (int, error) {
	ps, err := r.PorCliente(ctx, cliente)
	if err != nil {
		return 0, err
	}
	n := 0
	for _, p := range ps {
		if p.Estado == "pagado" {
			n++
		}
	}
	return n, nil
}
