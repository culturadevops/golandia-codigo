//go:build !malo

package pedidos

type Pedido struct {
	ID     int
	Estado string
}

// MarcarPagados pone en «pagado» los pedidos que confirmó el banco.
func MarcarPagados(ps []Pedido, pagados map[int]bool) {
	for i := range ps {
		if pagados[ps[i].ID] {
			ps[i].Estado = "pagado"
		}
	}
}
