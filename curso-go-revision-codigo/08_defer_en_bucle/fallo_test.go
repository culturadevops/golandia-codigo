package pedidos

import "testing"

func TestExportarDiezConPoolDeTres(t *testing.T) {
	ids := []int{101, 102, 103, 104, 105, 106, 107, 108, 109, 110}
	if err := Exportar(NuevoPool(3), ids); err != nil {
		t.Fatalf("10 pedidos, pool de 3 -> %v", err)
	}
	t.Logf("10 pedidos, pool de 3 -> exportados los 10")
}
