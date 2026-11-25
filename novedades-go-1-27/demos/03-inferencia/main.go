// Go 1.27: la inferencia de tipos de funciones genéricas se aplica en todos los contextos
// donde la función se asigna o se convierte a un tipo de función concreto.
package main

import (
	"fmt"
	"strings"
)

func Mayor[T int | float64 | string](a, b T) T {
	if a > b {
		return a
	}
	return b
}

type Regla struct {
	Nombre string
	Elegir func(a, b string) string
}

func main() {
	// Conversión explícita a un tipo de función.
	f := (func(int, int) int)(Mayor)
	fmt.Println("conversión:", f(3, 9))

	// Elemento de un literal de slice.
	ops := []func(float64, float64) float64{Mayor}
	fmt.Println("slice:", ops[0](1.5, 0.5))

	// Campo de un literal de struct.
	r := Regla{Nombre: "mayor", Elegir: Mayor}
	fmt.Println("struct:", r.Elegir("go", "rust"), strings.ToUpper(r.Nombre))
}
