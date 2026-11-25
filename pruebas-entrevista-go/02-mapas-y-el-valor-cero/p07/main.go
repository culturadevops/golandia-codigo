package main

import "fmt"

func main() {
	// delete sobre un mapa NIL: no explota. Al reves que escribir.
	var nulo map[string]int
	delete(nulo, "x")
	fmt.Println("delete sobre un mapa nil: no pasa nada")

	m := map[string]int{"a": 1, "b": 2, "c": 3}
	delete(m, "b")
	fmt.Println("tras borrar la b:", len(m), "claves")
	delete(m, "no-existe")
	fmt.Println("borrar algo que no esta tampoco falla:", len(m), "claves")

	// Borrar DURANTE el recorrido esta permitido explicitamente.
	g := map[int]int{}
	for i := 0; i < 10; i++ {
		g[i] = i
	}
	for k := range g {
		if k%2 == 0 {
			delete(g, k)
		}
	}
	fmt.Println("tras borrar los pares durante el range, quedan:", len(g))
}
