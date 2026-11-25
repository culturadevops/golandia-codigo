//go:build paso2

package fuzzing

import "fmt"

// Arreglado el fallo 1: el signo va delante y los centavos en positivo. -150 sale «-1,50»
// y -5 sale «-0,05». Negando cociente y resto por separado no se desborda ni con math.MinInt64.
func Formatear(m Monto) string {
	if m < 0 {
		return fmt.Sprintf("-%d,%02d", -(m / 100), -(m % 100))
	}
	return fmt.Sprintf("%d,%02d", m/100, m%100)
}
