//go:build !paso2

package fuzzing

import "fmt"

// Fallo 1: con un negativo, m%100 también es negativo. -150 sale «-1,-50».
func Formatear(m Monto) string {
	return fmt.Sprintf("%d,%02d", m/100, m%100)
}
