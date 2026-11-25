//go:build ignore

// 1b · Y asi es como se escribe de verdad. El fallo es una rama, no un salto.
package main

import (
	"errors"
	"fmt"
)

func cobrar(saldo, importe int) (int, error) {
	if importe > saldo {
		return 0, errors.New("saldo insuficiente")
	}
	return saldo - importe, nil
}

func main() {
	restante, err := cobrar(10, 50)
	if err != nil {
		fmt.Println("no se pudo cobrar:", err)
		return
	}
	fmt.Println("quedan:", restante)
}
