package main

import (
	"errors"
	"fmt"
	"strconv"
)

func leerEdad(s string) (int, error) {
	n, err := strconv.Atoi(s)
	if err != nil {
		return 0, fmt.Errorf("edad %q: %w", s, err)
	}
	if n < 0 {
		return 0, errors.New("la edad no puede ser negativa")
	}
	return n, nil
}

func main() {
	for _, s := range []string{"30", "treinta", "-5"} {
		edad, err := leerEdad(s)
		if err != nil {
			fmt.Println("error:", err)
			continue
		}
		fmt.Println("edad:", edad)
	}
}
