//go:build ignore

// 1 · El error es un valor que se DEVUELVE, y el compilador te obliga a recogerlo.
// El equivalente de la excepcion silenciosa de TypeScript aqui no compila.
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
	// La firma DICE que esto puede fallar: devuelve dos cosas.
	// Recoger solo una es un error de COMPILACION, no de produccion.
	restante := cobrar(10, 50)
	fmt.Println("quedan:", restante)
}
