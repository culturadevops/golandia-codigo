//go:build !malo

package pedidos

import (
	"strconv"
	"strings"
)

type Pedido struct {
	ID      int
	Cliente string
	Total   int
}

// CSV arma el reporte diario de pedidos.
// Un += sobre un string copia todo lo anterior en cada vuelta; el Builder no.
func CSV(ps []Pedido) string {
	var b strings.Builder
	b.Grow(32 * (len(ps) + 1)) // un cálculo a ojo del tamaño final: una reserva, no veinte
	b.WriteString("id,cliente,total\n")
	for _, p := range ps {
		b.WriteString(strconv.Itoa(p.ID))
		b.WriteByte(',')
		b.WriteString(p.Cliente)
		b.WriteByte(',')
		b.WriteString(strconv.Itoa(p.Total))
		b.WriteByte('\n')
	}
	return b.String()
}
