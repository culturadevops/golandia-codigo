// Go 1.27: vet avisa de fmt.Errorf("...: %w", p) cuando p es *E y E ya implementa error.
package main

import (
	"errors"
	"fmt"
)

// NoEncontrado implementa error con receptor de VALOR.
type NoEncontrado struct{ Clave string }

func (e NoEncontrado) Error() string { return "no encontrado: " + e.Clave }

func buscar(clave string) error {
	e := NoEncontrado{Clave: clave}
	return fmt.Errorf("buscar: %w", &e) // envuelve un *NoEncontrado
}

func main() {
	err := buscar("usuario:42")
	fmt.Println(err)

	var ne NoEncontrado
	fmt.Println("errors.As(err, &NoEncontrado):", errors.As(err, &ne))
}
