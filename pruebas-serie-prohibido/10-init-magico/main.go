// Video 211 — init() se ejecuta sin que la llames, y el orden entre ficheros
// NO es el que crees: es el ALFABETICO del nombre del fichero.
package main

import (
	"fmt"

	_ "initmagico/driver" // importado SOLO por su efecto: su init() corre igual
)

func init() { fmt.Println("  3. init() de main.go") }

func main() {
	fmt.Println("  4. y por fin main()")
	fmt.Println()
	fmt.Println("  Nadie llamo a ningun init(). Y no puedes llamarlos tu:")
	fmt.Println("  descomenta la linea de abajo y no compila.")
	// init()   // -> undefined: init
}
