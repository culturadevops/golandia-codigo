package main

import "fmt"

type Contador struct{ N int }

func (c Contador) SumaValor()   { c.N++ } // recibe una COPIA
func (c *Contador) SumaPuntero() { c.N++ }

func main() {
	c := Contador{}
	for i := 0; i < 3; i++ {
		c.SumaValor()
	}
	fmt.Println("tras 3 SumaValor():  ", c.N, "  <-- no cambio nada")

	for i := 0; i < 3; i++ {
		c.SumaPuntero()
	}
	fmt.Println("tras 3 SumaPuntero():", c.N)
}
