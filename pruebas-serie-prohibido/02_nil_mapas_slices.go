package main

import "fmt"

type MiError struct{}

func (e *MiError) Error() string { return "algo fallo" }

// Devuelve un *MiError que vale nil... metido en una interfaz error.
func puedeFallar(falla bool) error {
	var e *MiError
	if falla {
		e = &MiError{}
	}
	return e // <- AQUI esta la trampa
}

func main() {
	fmt.Println("=== 1 · nil que no es nil ===")
	err := puedeFallar(false)
	fmt.Println("  err ==", err)
	fmt.Println("  err == nil ?", err == nil, "  <- todo el mundo espera true")

	fmt.Println("\n=== 2 · los mapas se recorren en orden ALEATORIO a proposito ===")
	m := map[string]int{"a": 1, "b": 2, "c": 3, "d": 4, "e": 5}
	for vuelta := 1; vuelta <= 3; vuelta++ {
		fmt.Print("  vuelta ", vuelta, ": ")
		for k := range m {
			fmt.Print(k, " ")
		}
		fmt.Println()
	}

	fmt.Println("\n=== 3 · dos slices comparten memoria y se pisan ===")
	original := []int{1, 2, 3, 4, 5}
	trozo := original[1:3]
	fmt.Println("  original :", original)
	fmt.Println("  trozo    :", trozo, " (len", len(trozo), "cap", cap(trozo), ")")
	trozo = append(trozo, 99)
	fmt.Println("  tras append(trozo, 99):")
	fmt.Println("  trozo    :", trozo)
	fmt.Println("  original :", original, " <- CAMBIO SOLO, sin tocarlo")
}
