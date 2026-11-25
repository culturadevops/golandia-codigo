package main

import "fmt"

func main() {
	origen := []int{1, 2, 3, 4, 5}

	var destino []int
	n := copy(destino, origen)
	fmt.Println("copy a un slice nil:", n, "copiados", destino)

	destino = make([]int, 3)
	n = copy(destino, origen)
	fmt.Println("copy a make(3):     ", n, "copiados", destino)

	destino = make([]int, len(origen))
	n = copy(destino, origen)
	destino[0] = 99
	fmt.Println("copia completa:     ", n, "copiados", destino, "origen", origen)
}
