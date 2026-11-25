package main

import "fmt"

type Figura interface {
	Area() float64
	Perimetro() float64
}

type Rectangulo struct{ Ancho, Alto float64 }

func (r Rectangulo) Area() float64 { return r.Ancho * r.Alto }

func main() {
	var f Figura = Rectangulo{3, 4}
	fmt.Println(f.Area())
}
