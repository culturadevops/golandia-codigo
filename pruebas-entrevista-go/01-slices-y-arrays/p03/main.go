package main

import "fmt"

func main() {
	base := []int{1, 2, 3, 4, 5}

	// Corte de tres indices: [inicio:fin:tope]. El tope recorta la CAPACIDAD.
	seguro := base[1:3:3]
	fmt.Println("base[1:3:3] -> len", len(seguro), "cap", cap(seguro))
	seguro = append(seguro, 99)
	fmt.Println("tras append -> base:", base, " (intacta)")

	// La otra salida: copy(), que no comparte nada desde el primer momento.
	copia := make([]int, 2)
	n := copy(copia, base[1:3])
	copia = append(copia, 99)
	fmt.Println("copy copio", n, "elementos -> copia:", copia, " base:", base)
}
