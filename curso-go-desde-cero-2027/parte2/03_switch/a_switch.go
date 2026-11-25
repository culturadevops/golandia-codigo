package main

import "fmt"

func dia(n int) string {
	switch n {
	case 1, 2, 3, 4, 5:
		return "laborable"
	case 6, 7:
		return "fin de semana"
	default:
		return "no existe"
	}
}

func nota(x int) string {
	switch {
	case x >= 90:
		return "excelente"
	case x >= 70:
		return "aprobado"
	}
	return "a repasar"
}

func main() {
	fmt.Println(dia(3), dia(7), dia(9))
	fmt.Println(nota(95), nota(72), nota(10))
}
