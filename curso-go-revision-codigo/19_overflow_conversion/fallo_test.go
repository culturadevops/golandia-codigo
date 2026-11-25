package pedidos

import "testing"

func TestCantidadQueNoCabe(t *testing.T) {
	n, err := Cantidad("40000")
	if err == nil {
		t.Fatalf(`Cantidad("40000") = %d, err = nil: dio la vuelta y quedó negativa`, n)
	}
	t.Logf(`Cantidad("40000") -> %v`, err)
}
