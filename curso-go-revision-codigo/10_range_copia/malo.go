//go:build malo

package pedidos

type Pedido struct {
	ID     int
	Estado string
}

// MarcarPagados pone en «pagado» los pedidos que confirmó el banco.
func MarcarPagados(ps []Pedido, pagados map[int]bool) {
	for _, p := range ps {
		if pagados[p.ID] {
			p.Estado = "pagado"
		}
	}
}
