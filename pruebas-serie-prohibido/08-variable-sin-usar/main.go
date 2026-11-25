// Video 203 — Go NO COMPILA por una variable declarada y no usada.
// El titular es LITERAL: no es un aviso, es un error. Esto no compila.
package main

import (
	"fmt"
	"os" // <- importado y no usado: TAMBIEN es error
)

func main() {
	contador := 0 // <- declarada y no usada: error
	fmt.Println("hola")
}
