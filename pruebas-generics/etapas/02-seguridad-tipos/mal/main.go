package main

import (
	"fmt"

	"seguridad/comun"
)

func main() {
	palabras := []string{"a", "b", "c"}
	// string no cumple la restriccion Number: esto no debe compilar
	fmt.Println(comun.SumGeneric(palabras))
}
