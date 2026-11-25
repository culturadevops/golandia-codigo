package main

import (
	"fmt"

	"seguridad/comun"
)

func main() {
	enteros := []int{10, 20, 30}
	fmt.Println("SumGeneric con []int:", comun.SumGeneric(enteros))

	flotantes := []float64{1.5, 2.5}
	fmt.Println("SumGeneric con []float64:", comun.SumGeneric(flotantes))

	// un dato equivocado se cuela como interface{}, y SumInterface lo descarta EN SILENCIO
	mixto := []interface{}{10, 20, "treinta"}
	fmt.Println("SumInterface con un string colado:", comun.SumInterface(mixto), "(debería ser un error, no 30)")
}
