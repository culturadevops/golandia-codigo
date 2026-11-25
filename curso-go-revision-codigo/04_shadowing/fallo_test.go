package pedidos

import "testing"

func TestTotalConUnaLineaMala(t *testing.T) {
	total, err := Total([]string{"10", "x", "5"})
	if err == nil {
		t.Fatalf(`Total("10", "x", "5") = %d, err = nil: el error se quedó dentro del bucle`, total)
	}
	t.Logf(`Total("10", "x", "5") -> %v`, err)
}
