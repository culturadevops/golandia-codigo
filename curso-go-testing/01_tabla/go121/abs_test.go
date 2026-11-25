// Un módulo aparte con «go 1.21» en su go.mod. Con esa versión del lenguaje, la variable del
// for es UNA para todas las vueltas: los subtests paralelos se ejecutan cuando el bucle ya
// terminó y todos ven el último caso. El caso que falla nunca se prueba.
//
//	cd 01_tabla/go121 && go test -v .        -> todo verde (y es mentira)
//
// Cambia la línea «go 1.21» por «go 1.22» y vuelve a correrlo.
package go121

import "testing"

// Abs tiene un fallo: con un negativo devuelve el negativo.
func Abs(n int) int {
	if n < 0 {
		return n // debería ser -n
	}
	return n
}

func TestAbs(t *testing.T) {
	casos := []struct {
		nombre     string
		entra, sale int
	}{
		{"positivo", 3, 3},
		{"negativo", -3, 3},
		{"cero", 0, 0},
	}
	for _, tc := range casos {
		t.Run(tc.nombre, func(t *testing.T) {
			t.Parallel()
			t.Logf("pruebo el caso %q: Abs(%d)", tc.nombre, tc.entra)
			if got := Abs(tc.entra); got != tc.sale {
				t.Errorf("Abs(%d) = %d; quiero %d", tc.entra, got, tc.sale)
			}
		})
	}
}
