package main

import "fmt"

// El caso real: una funcion que hace append y NO devuelve el slice.
func agregar(s []int, v int) {
	s = append(s, v)
	fmt.Println("  dentro de agregar:", s)
}

func main() {
	s := make([]int, 0, 4)
	s = append(s, 1, 2)
	fmt.Println("antes:", s, "len", len(s), "cap", cap(s))
	agregar(s, 3)
	fmt.Println("despues:", s, "  <-- el 3 no esta: len sigue siendo", len(s))
	fmt.Println("pero SI se escribio en el array de debajo:", s[:cap(s)])
}
