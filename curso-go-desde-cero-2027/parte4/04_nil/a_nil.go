package main

import "fmt"

type ErrorValidacion struct {
	Campo string
}

func (e *ErrorValidacion) Error() string { return "campo inválido: " + e.Campo }

func validar(nombre string) error {
	var err *ErrorValidacion
	if nombre == "" {
		err = &ErrorValidacion{Campo: "nombre"}
	}
	return err
}

func main() {
	err := validar("Ana")
	fmt.Printf("err = %v · tipo %T\n", err, err)
	if err != nil {
		fmt.Println("¿err != nil? sí: la validación falló... con un nombre válido")
	}
}
