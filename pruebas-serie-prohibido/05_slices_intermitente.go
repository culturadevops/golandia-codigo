// 05_slices_intermitente.go — por que el pisotón entre slices es INTERMITENTE.
//
// POR QUE EXISTE. 02_nil_mapas_slices.go demuestra que `append` sobre un trozo pisa el original.
// Pero eso solo pasa cuando SOBRA capacidad prestada. Cuando no sobra, `append` copia a un array
// nuevo y el original NO se toca. Ese es el fallo de verdad: el mismo codigo corrompe datos o no
// segun cuanta capacidad quedara, o sea segun el tamano de la entrada.
//
// Un fallo que siempre pasa se encuentra. Uno que depende del tamano de la entrada llega a
// produccion.
//
// Correr:  go run 05_slices_intermitente.go
package main

import "fmt"

func prueba(nombre string, original []int, desde, hasta int) {
	copia := append([]int{}, original...) // para poder comparar despues
	trozo := original[desde:hasta]
	fmt.Printf("  %-22s original=%v  trozo=%v  len=%d cap=%d\n",
		nombre, original, trozo, len(trozo), cap(trozo))
	trozo = append(trozo, 99)
	pisado := false
	for i := range original {
		if original[i] != copia[i] {
			pisado = true
		}
	}
	estado := "intacto"
	if pisado {
		estado = "PISADO"
	}
	fmt.Printf("  %-22s tras append(trozo,99): original=%v  -> %s\n\n", "", original, estado)
}

func main() {
	fmt.Println("=== el MISMO codigo, dos entradas distintas ===")
	fmt.Println()

	// Sobra capacidad: el trozo llega hasta el indice 3 de 5, quedan 2 huecos.
	prueba("sobra capacidad", []int{1, 2, 3, 4, 5}, 1, 3)

	// NO sobra: el trozo llega hasta el final, no queda hueco -> append copia a otro array.
	prueba("capacidad justa", []int{1, 2, 3}, 1, 3)

	fmt.Println("=== y lo mismo, barriendo tamanos ===")
	for n := 3; n <= 8; n++ {
		orig := make([]int, n)
		for i := range orig {
			orig[i] = i + 1
		}
		copia := append([]int{}, orig...)
		trozo := orig[1:3]
		c := cap(trozo)
		trozo = append(trozo, 99)
		pisado := false
		for i := range orig {
			if orig[i] != copia[i] {
				pisado = true
			}
		}
		fmt.Printf("  len(original)=%d  cap(trozo)=%d  ->  %s\n", n, c,
			map[bool]string{true: "PISADO", false: "intacto"}[pisado])
	}
	fmt.Println()
	fmt.Println("El fallo depende del TAMANO DE LA ENTRADA. Por eso llega a produccion.")
}
