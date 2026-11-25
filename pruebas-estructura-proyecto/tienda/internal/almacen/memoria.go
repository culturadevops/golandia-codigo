// Package almacen guarda pedidos. Hoy en memoria; mañana Postgres, sin tocar pedidos.
package almacen

import (
	"sync"

	"github.com/golandia/tienda/internal/pedidos"
)

type Memoria struct {
	mu    sync.Mutex
	datos []pedidos.Pedido
}

func (m *Memoria) Guardar(p pedidos.Pedido) (int, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.datos = append(m.datos, p)
	return len(m.datos), nil
}
