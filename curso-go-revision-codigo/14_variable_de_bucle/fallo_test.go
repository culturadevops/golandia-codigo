package pedidos

import (
	"fmt"
	"testing"
)

func TestUnAvisoPorPedido(t *testing.T) {
	var salen []string
	for _, aviso := range Avisos([]int{7, 8, 9}) {
		salen = append(salen, aviso())
	}
	if fmt.Sprint(salen) != "[pedido 7 listo pedido 8 listo pedido 9 listo]" {
		t.Fatalf("avisos de 7, 8 y 9 -> %q", salen)
	}
	t.Logf("avisos de 7, 8 y 9 -> %q", salen)
}
