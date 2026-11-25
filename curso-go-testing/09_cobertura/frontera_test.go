//go:build trampa

package cobertura

import "testing"

// El caso que falta: justo en la frontera. Con él la cobertura no sube (ya era 100 %),
// pero el test por fin dice la verdad.
func TestEnvioFrontera(t *testing.T) {
	for _, c := range []struct {
		sub, want Monto
	}{{49999, Costo}, {50000, 0}, {50001, 0}} {
		if got := Envio(c.sub); got != c.want {
			t.Errorf("Envio(%d) = %d; quiero %d", c.sub, got, c.want)
		}
	}
}
