// El mismo descuento de los lunes, con el mismo error, en Go.
package main

import "fmt"

func precioFinal(precio int, dia string) int {
	if dia == "lunes" {
		return precio - "10" // error: resta un texto a un número
	}
	return precio
}

func main() {
	fmt.Println("martes:", precioFinal(100, "martes"))
	fmt.Println("lunes:", precioFinal(100, "lunes"))
}
