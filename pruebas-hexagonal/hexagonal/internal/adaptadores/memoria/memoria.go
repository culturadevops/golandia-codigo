// Package memoria es un ADAPTADOR del puerto pedidos.Repositorio que guarda en un mapa. Sirve para las pruebas del dominio.
package memoria

import (
	"context"
	"sync"

	"tienda/internal/pedidos"
)

type Repo struct {
	mu    sync.Mutex
	datos map[int64]pedidos.Pedido
	sig   int64
}

func Nuevo() *Repo { return &Repo{datos: map[int64]pedidos.Pedido{}} }

func (r *Repo) Guardar(_ context.Context, p pedidos.Pedido) (int64, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.sig++
	p.ID = r.sig
	r.datos[p.ID] = p
	return p.ID, nil
}

func (r *Repo) Buscar(_ context.Context, id int64) (pedidos.Pedido, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	p, ok := r.datos[id]
	if !ok {
		return pedidos.Pedido{}, pedidos.ErrNoEncontrado
	}
	return p, nil
}
