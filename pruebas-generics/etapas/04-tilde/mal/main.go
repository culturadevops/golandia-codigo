package main

import (
	"fmt"

	"tilde/comun"
)

func main() {
	temperaturas := []comun.Celsius{20.5, 21.0, 19.5}
	// Celsius tiene tipo subyacente float64, pero SinTilde exige float64 EXACTO: esto no compila
	fmt.Println(comun.SumSinTilde(temperaturas))
}
