package main

import "fmt"

func main() {
	var nulo []int          // slice NIL: nunca se le asigno nada
	vacio := []int{}        // slice vacio, pero existe

	fmt.Println("nulo  == nil?", nulo == nil, " len", len(nulo), "cap", cap(nulo))
	fmt.Println("vacio == nil?", vacio == nil, "len", len(vacio), "cap", cap(vacio))
	fmt.Println()
	fmt.Println("Y aun asi, los dos hacen lo mismo:")
	nulo = append(nulo, 1)  // append sobre nil FUNCIONA
	fmt.Println("append sobre el nil ->", nulo)
	for range nulo[:0] {
		fmt.Println("nunca")
	}
	fmt.Println("recorrer un slice nil tampoco explota")
}
