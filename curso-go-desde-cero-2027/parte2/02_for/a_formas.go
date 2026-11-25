package main

import "fmt"

func main() {
	// 1. clásico
	for i := 0; i < 3; i++ {
		fmt.Print(i, " ")
	}
	fmt.Println("<- clásico")

	// 2. como un while
	n := 1
	for n < 100 {
		n *= 3
	}
	fmt.Println(n, "<- como while")

	// 3. infinito con break
	intentos := 0
	for {
		intentos++
		if intentos == 4 {
			break
		}
	}
	fmt.Println(intentos, "<- infinito con break")

	// 4. range sobre un número (Go 1.22+)
	for i := range 3 {
		fmt.Print(i, " ")
	}
	fmt.Println("<- range 3")

	// 5. range sobre un slice, y continue
	for i, f := range []string{"mango", "guayaba", "parchita"} {
		if i == 1 {
			continue
		}
		fmt.Println(i, f)
	}
}
