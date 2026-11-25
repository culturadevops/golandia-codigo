package main

import "fmt"

// El MISMO append. Lo unico que cambia es si cabia en la capacidad.
func main() {
	for _, corte := range [][2]int{{1, 3}, {3, 5}} {
		base := []int{1, 2, 3, 4, 5}
		t := base[corte[0]:corte[1]]
		fmt.Printf("base[%d:%d] -> len %d cap %d ", corte[0], corte[1], len(t), cap(t))
		antes := fmt.Sprint(base)
		t = append(t, 99)
		despues := fmt.Sprint(base)
		if antes == despues {
			fmt.Printf("| append REASIGNO, base intacta: %s\n", despues)
		} else {
			fmt.Printf("| append PISO en sitio: %s -> %s\n", antes, despues)
		}
	}
}
