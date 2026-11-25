// Dos comprobaciones para el video 211:
//   1. ¿se pueden poner VARIAS init() en el mismo fichero?
//   2. ¿que pasa si intentas llamarla tu?
package main

import "fmt"

func init() { fmt.Println("  init() numero 1 del mismo fichero") }
func init() { fmt.Println("  init() numero 2 del mismo fichero") }
func init() { fmt.Println("  init() numero 3 del mismo fichero") }

func main() {
	fmt.Println("  main()")
}
