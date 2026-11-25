package main

import "fmt"

type Figura interface {
	Area() float64
}

type Cuadrado struct{ Lado float64 }

func (c *Cuadrado) Area() float64 { return c.Lado * c.Lado }

func main() {
	var f Figura = Cuadrado{2}
	fmt.Println(f.Area())
}
