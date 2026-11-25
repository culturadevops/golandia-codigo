package main

import "fmt"

type Informe struct {
	Titulo string
	Filas  []int
}

// Recibe el struct POR VALOR. Se copia entero... o eso parece.
func tocar(inf Informe) {
	inf.Titulo = "cambiado"  // esto no sale de aqui
	inf.Filas[0] = 999       // esto SI sale
}

func main() {
	inf := Informe{Titulo: "original", Filas: []int{1, 2, 3}}
	tocar(inf)
	fmt.Println("Titulo:", inf.Titulo, " <- la copia protegio el campo")
	fmt.Println("Filas: ", inf.Filas, "  <- pero NO el contenido del slice")
	fmt.Println()
	fmt.Println("En Go TODO se pasa por valor. Lo que se copio fue la cabecera")
	fmt.Println("del slice, y esa cabecera apunta al mismo array.")
}
