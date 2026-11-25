package main

import "fmt"

type Punto struct{ X, Y int }

// Se queda en la PILA: no sale de la funcion.
func enPila() int {
	p := Punto{1, 2}
	return p.X + p.Y
}

// Se ESCAPA al monton: se devuelve un puntero a algo local.
func seEscapa() *Punto {
	p := Punto{1, 2}
	return &p
}

// Tambien se escapa, y esta sorprende: fmt.Println recibe `any`, y eso obliga
// a que el valor viva en el monton aunque no salga de aqui.
func porCulpaDeFmt() {
	p := Punto{3, 4}
	fmt.Println(p.X)
	_ = p
}

func main() {
	_ = enPila()
	_ = seEscapa()
	porCulpaDeFmt()
}
