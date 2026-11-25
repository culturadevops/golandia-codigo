//go:build !malo

package pedidos

import "sync"

type Inventario struct {
	mu    sync.Mutex
	stock map[string]int
}

func NuevoInventario() *Inventario { return &Inventario{stock: map[string]int{}} }

// Vender la llaman a la vez todos los handlers del servicio.
func (i *Inventario) Vender(sku string) {
	i.mu.Lock()
	defer i.mu.Unlock()
	i.stock[sku]--
}

func (i *Inventario) Stock(sku string) int {
	i.mu.Lock()
	defer i.mu.Unlock()
	return i.stock[sku]
}
