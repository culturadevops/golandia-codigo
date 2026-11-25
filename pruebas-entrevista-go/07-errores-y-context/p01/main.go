package main

import (
	"errors"
	"fmt"
)

var ErrNoEncontrado = errors.New("no encontrado")

func buscarBien(id int) error {
	return fmt.Errorf("buscando usuario %d: %w", id, ErrNoEncontrado) // %w
}

func buscarMal(id int) error {
	return fmt.Errorf("buscando usuario %d: %v", id, ErrNoEncontrado) // %v
}

func main() {
	bien, mal := buscarBien(7), buscarMal(7)

	fmt.Println("con %w ->", bien)
	fmt.Println("con %v ->", mal)
	fmt.Println("   los dos mensajes son IDENTICOS. Esa es la trampa.")
	fmt.Println()
	fmt.Println("errors.Is(bien, ErrNoEncontrado):", errors.Is(bien, ErrNoEncontrado))
	fmt.Println("errors.Is(mal,  ErrNoEncontrado):", errors.Is(mal, ErrNoEncontrado), " <-- la cadena esta rota")
	fmt.Println()
	fmt.Println("Unwrap(bien):", errors.Unwrap(bien))
	fmt.Println("Unwrap(mal): ", errors.Unwrap(mal))
}
