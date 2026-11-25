package main

import "fmt"

type Punto struct{ X, Y int }

func main() {
	// Un struct SI puede ser clave: es comparable.
	visitados := map[Punto]bool{}
	visitados[Punto{1, 2}] = true
	fmt.Println("struct como clave:", visitados[Punto{1, 2}])

	// Un array tambien.
	porPar := map[[2]int]string{{1, 2}: "uno-dos"}
	fmt.Println("array como clave:", porPar[[2]int{1, 2}])

	fmt.Println("lo que NO puede ser clave: slices, mapas y funciones")
}
