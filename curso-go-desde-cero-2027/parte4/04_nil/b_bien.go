package main

import "fmt"

type ErrorValidacion struct {
	Campo string
}

func (e *ErrorValidacion) Error() string { return "campo inválido: " + e.Campo }

func validar(nombre string) error {
	if nombre == "" {
		return &ErrorValidacion{Campo: "nombre"}
	}
	return nil
}

func main() {
	fmt.Println("validar(\"Ana\") =", validar("Ana"))
	fmt.Println("validar(\"\")    =", validar(""))
}
