package main

import "fmt"

func main() {
	var m map[string]int // NIL. Nunca se hizo make.

	fmt.Println("m == nil?", m == nil)
	fmt.Println("len(m):", len(m))
	fmt.Println("leer una clave que no existe:", m["loquesea"], "<- no explota, devuelve el cero")

	edades := map[string]int{"ana": 34}
	fmt.Println("edades[pedro]:", edades["pedro"], "<- pedro no existe y aun asi hay un numero")
}
