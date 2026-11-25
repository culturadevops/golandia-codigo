//go:build malo

package pedidos

import "strconv"

type Pedido struct {
	ID      int
	Cliente string
	Total   int
}

// CSV arma el reporte diario de pedidos.
func CSV(ps []Pedido) string {
	s := "id,cliente,total\n"
	for _, p := range ps {
		s += strconv.Itoa(p.ID) + "," + p.Cliente + "," + strconv.Itoa(p.Total) + "\n"
	}
	return s
}
