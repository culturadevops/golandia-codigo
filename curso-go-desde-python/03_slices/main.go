// Leccion 3 — Slices: la lista que a veces comparte memoria y a veces no.
package main

import "fmt"

func main() {
	fmt.Println("== 3.1 · Un trozo de un slice es una VENTANA, no una copia ==")
	hosts := []string{"web-01", "web-02", "web-03", "web-04"}
	primeros := hosts[:2]
	primeros[0] = "CAMBIADO"
	fmt.Printf("  hosts    = %v\n", hosts)
	fmt.Printf("  primeros = %v\n", primeros)
	fmt.Println("  Cambiaste el trozo y cambio el original. En Python, hosts[:2]")
	fmt.Println("  es una COPIA: ahi no pasaria.")

	fmt.Println()
	fmt.Println("== 3.2 · LA TRAMPA: append a veces pisa al vecino ==")
	base := make([]string, 3, 10)
	copy(base, []string{"a", "b", "c"})
	x := base[:2]
	x = append(x, "X")
	fmt.Printf("  base = %v   <- el append de x piso base[2]\n", base)
	fmt.Printf("  len(base)=%d  cap(base)=%d\n", len(base), cap(base))
	fmt.Println("  Habia capacidad libre, asi que append escribio encima en vez de")
	fmt.Println("  crear un array nuevo. Si no hubiera habido capacidad, no pisa.")
	fmt.Println("  El mismo codigo, dos comportamientos, segun la capacidad.")

	fmt.Println()
	fmt.Println("== 3.3 · Copiar de verdad ==")
	orig := []string{"web-01", "web-02"}
	copia := make([]string, len(orig))
	copy(copia, orig)
	copia[0] = "otro"
	fmt.Printf("  orig=%v  copia=%v\n", orig, copia)
}
