package main

import "fmt"

type MiError struct{ Codigo int }

func (e *MiError) Error() string { return fmt.Sprintf("fallo %d", e.Codigo) }

// LA CORRECCION: no declarar la variable con tipo concreto para devolverla como error.
func bien(fallar bool) error {
	if fallar {
		return &MiError{Codigo: 500}
	}
	return nil // nil LITERAL: las dos palabras vacias
}

func main() {
	for _, fallar := range []bool{false, true} {
		err := bien(fallar)
		fmt.Printf("fallar=%-5v  err=%-9v  err != nil -> %v\n", fallar, err, err != nil)
	}
}
