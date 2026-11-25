// Lección 1 · Tests de tabla, subtests, t.Run y t.Parallel.
package tabla

import (
	"testing"

	"curso-testing/tienda"
)

func TestTotal(t *testing.T) {
	t.Parallel()
	taza := tienda.Item{SKU: "TAZA", Cantidad: 1, Precio: 15000}
	casos := []struct {
		nombre  string
		items   []tienda.Item
		primero bool
		want    tienda.Monto
	}{
		{"una taza, cliente nuevo", []tienda.Item{taza}, true, 15000 - 1500 + 4900},
		{"una taza, cliente de siempre", []tienda.Item{taza}, false, 15000 + 4900},
		{"justo 500,00: envío gratis", []tienda.Item{{SKU: "MOLINO", Cantidad: 1, Precio: 50000}}, false, 50000},
		{"499,99: paga envío", []tienda.Item{{SKU: "MOLINO", Cantidad: 1, Precio: 49999}}, false, 49999 + 4900},
		{"cantidades", []tienda.Item{{SKU: "TAZA", Cantidad: 3, Precio: 15000}, {SKU: "FILTRO", Cantidad: 10, Precio: 2000}}, false, 65000},
		{"descuento redondea hacia abajo", []tienda.Item{{SKU: "CHICLE", Cantidad: 1, Precio: 99}}, true, 99 - 9 + 4900},
	}
	for _, tc := range casos {
		t.Run(tc.nombre, func(t *testing.T) {
			t.Parallel()
			if got := tienda.Total(tc.items, tc.primero); got != tc.want {
				t.Errorf("Total = %s; quiero %s", got, tc.want)
			}
		})
	}
}

// Cada subtest se puede correr solo: go test -run 'TestTotal/499' ./01_tabla
// (los espacios del nombre se convierten en guiones bajos).
