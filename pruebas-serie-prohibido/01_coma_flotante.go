package main

import "fmt"

func main() {
	fmt.Println("--- Go tiene EXACTAMENTE el mismo problema ---")
	fmt.Println("0.1 + 0.2      =", 0.1+0.2)
	fmt.Println("¿es igual a 0.3?", 0.1+0.2 == 0.3)

	// El caso real: un carrito de la compra
	var total float64
	for i := 0; i < 10; i++ {
		total += 0.1
	}
	fmt.Println("\n10 veces 0,10 € =", total)
	fmt.Println("¿es 1 euro?     ", total == 1.0)

	// Un precio de tienda
	precio := 19.99
	cantidad := 3.0
	fmt.Printf("\n19,99 x 3      = %.20f\n", precio*cantidad)

	// LA SOLUCION: centimos en enteros
	fmt.Println("\n--- en céntimos (enteros) ---")
	var centimos int64
	for i := 0; i < 10; i++ {
		centimos += 10
	}
	fmt.Println("10 veces 10 céntimos =", centimos, "céntimos =", float64(centimos)/100, "€")
	fmt.Println("¿exacto?", centimos == 100)
}
