//go:build malo

package pedidos

type Inventario struct {
	stock map[string]int
}

func NuevoInventario() *Inventario { return &Inventario{stock: map[string]int{}} }

// Vender la llaman a la vez todos los handlers del servicio.
func (i *Inventario) Vender(sku string) { i.stock[sku]-- }

func (i *Inventario) Stock(sku string) int { return i.stock[sku] }
