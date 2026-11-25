// Video 206 — por que la "solucion" con genericos NO es un ternario.
//
// Un ternario de verdad (C, Java, JS) es PEREZOSO: solo evalua la rama que gana.
// Una funcion de Go evalua SIEMPRE sus dos argumentos antes de entrar.
package main

import "fmt"

func Si[T any](cond bool, siSi, siNo T) T {
	if cond {
		return siSi
	}
	return siNo
}

func caro(nombre string, valor int) int {
	fmt.Println("      ...evaluando", nombre)
	return valor
}

func main() {
	fmt.Println("=== la funcion generica que todos proponen ===")
	r := Si(true, caro("rama VERDADERA", 1), caro("rama FALSA", 2))
	fmt.Println("  resultado:", r)
	fmt.Println("  -> se evaluaron LAS DOS, aunque solo hacia falta una")

	fmt.Println()
	fmt.Println("=== y con un if de toda la vida ===")
	var r2 int
	if true {
		r2 = caro("rama VERDADERA", 1)
	} else {
		r2 = caro("rama FALSA", 2)
	}
	fmt.Println("  resultado:", r2)
	fmt.Println("  -> solo se evaluo la que hacia falta")

	fmt.Println()
	fmt.Println("=== y por eso esto REVIENTA con la funcion, y no con el if ===")
	var lista []int // nil, longitud 0
	defer func() {
		if p := recover(); p != nil {
			fmt.Println("  panic:", p)
			fmt.Println("  -> la rama 'lista[0]' se evaluo AUNQUE la condicion decia que no")
		}
	}()
	fmt.Println("  Si(len(lista) > 0, lista[0], -1)")
	_ = Si(len(lista) > 0, lista[0], -1)
}
