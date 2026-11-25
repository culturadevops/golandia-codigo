package main

import (
	"fmt"
	"math"
)

type Figura interface {
	Area() float64
}

type Rectangulo struct{ Ancho, Alto float64 }
type Circulo struct{ Radio float64 }

func (r Rectangulo) Area() float64 { return r.Ancho * r.Alto }
func (c Circulo) Area() float64    { return math.Pi * c.Radio * c.Radio }

func total(fs []Figura) float64 {
	suma := 0.0
	for _, f := range fs {
		suma += f.Area()
	}
	return suma
}

func main() {
	figuras := []Figura{Rectangulo{3, 4}, Circulo{1}}
	for _, f := range figuras {
		fmt.Printf("%T: %.2f\n", f, f.Area())
	}
	fmt.Printf("total: %.2f\n", total(figuras))
}
