package main

import (
	"errors"
	"fmt"
)

var ErrSinStock = errors.New("sin stock")

func reservar(producto string, cantidad int) error {
	if cantidad > 3 {
		return ErrSinStock
	}
	return nil
}

func comprarConW(cantidad int) error {
	if err := reservar("café", cantidad); err != nil {
		return fmt.Errorf("comprar %d cafés: %w", cantidad, err)
	}
	return nil
}

func comprarConV(cantidad int) error {
	if err := reservar("café", cantidad); err != nil {
		return fmt.Errorf("comprar %d cafés: %v", cantidad, err)
	}
	return nil
}

func main() {
	e1 := comprarConW(5)
	e2 := comprarConV(5)
	fmt.Println("con %w:", e1)
	fmt.Println("con %v:", e2)
	fmt.Println("errors.Is(con %w, ErrSinStock):", errors.Is(e1, ErrSinStock))
	fmt.Println("errors.Is(con %v, ErrSinStock):", errors.Is(e2, ErrSinStock))
	fmt.Println("e1 == ErrSinStock:", e1 == ErrSinStock)
	fmt.Println("comprar 2:", comprarConW(2))
}
