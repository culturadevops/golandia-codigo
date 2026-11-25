package pedidos

import "testing"

func TestMarcarPagados(t *testing.T) {
	ps := []Pedido{{1, "pendiente"}, {2, "pendiente"}, {3, "pendiente"}}
	MarcarPagados(ps, map[int]bool{1: true, 3: true})
	if ps[0].Estado != "pagado" || ps[2].Estado != "pagado" {
		t.Fatalf("el banco confirmó 1 y 3 -> %v: se cambió la copia, no el pedido", ps)
	}
	t.Logf("el banco confirmó 1 y 3 -> %v", ps)
}
