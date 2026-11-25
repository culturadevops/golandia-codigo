package pedidos

import "testing"

func TestReciboEnJSON(t *testing.T) {
	b, err := NuevoRecibo(7, 12.5).JSON()
	if string(b) != `{"id":7,"total":12.5}` {
		t.Fatalf("recibo del pedido 7 -> %s (err = %v)", b, err)
	}
	t.Logf("recibo del pedido 7 -> %s", b)
}
