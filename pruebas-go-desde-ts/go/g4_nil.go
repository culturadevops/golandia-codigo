//go:build ignore

// 4 · nil no es null, y no se comporta igual segun el tipo. Esta asimetria es la que muerde.
package main

import "fmt"

func rescatar(que string) {
	if r := recover(); r != nil {
		fmt.Printf("%-28s PANIC: %v\n", que, r)
	}
}

func main() {
	// mapa nil: LEER esta bien, ESCRIBIR revienta.
	var m map[string]int
	fmt.Printf("%-28s %d (y no revienta)\n", "leer de un mapa nil:", m["loquesea"])
	func() {
		defer rescatar("escribir en un mapa nil:")
		m["algo"] = 1
	}()

	// slice nil: se le puede anadir sin inicializar nada.
	var s []int
	s = append(s, 1, 2, 3)
	fmt.Printf("%-28s %v (largo %d)\n", "append a un slice nil:", s, len(s))

	// puntero nil: aqui si.
	type Cliente struct{ Nombre string }
	var p *Cliente
	func() {
		defer rescatar("leer un puntero nil:")
		fmt.Println(p.Nombre)
	}()
}
