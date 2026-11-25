// Código de un equipo que usa la biblioteca: llama a la función vieja.
package main

import (
	"fmt"

	"demo/viejo"
)

func main() {
	fmt.Println(viejo.Limpiar("  ANA@Ejemplo.com "))
	fmt.Println(viejo.Limpiar(" Luis "))
}
