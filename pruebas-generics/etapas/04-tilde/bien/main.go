package main

import (
	"fmt"

	"tilde/comun"
)

func main() {
	temperaturas := []comun.Celsius{20.5, 21.0, 19.5}
	fmt.Println("SumConTilde con []Celsius:", comun.SumConTilde(temperaturas))

	planas := []float64{1.5, 2.5}
	fmt.Println("SumConTilde con []float64:", comun.SumConTilde(planas))
	fmt.Println("SumSinTilde con []float64:", comun.SumSinTilde(planas))
}
