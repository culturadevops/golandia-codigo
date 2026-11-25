package pedidos

import "testing"

func TestPedidoQueNoExisteDa404(t *testing.T) {
	if c := Codigo(99); c != 404 {
		t.Fatalf("GET /pedidos/99 -> %d; esperaba 404 (el error va envuelto con %%w)", c)
	}
	t.Logf("GET /pedidos/99 -> 404")
}
