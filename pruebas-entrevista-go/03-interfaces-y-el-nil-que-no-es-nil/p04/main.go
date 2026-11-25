package main

import (
	"errors"
	"fmt"
)

type MiError struct{ Codigo int }

func (e *MiError) Error() string { return fmt.Sprintf("fallo %d", e.Codigo) }

func malo() error {
	var err *MiError
	return err
}

func main() {
	err := malo()

	// La repregunta: y errors.As, que hace con esto?
	var destino *MiError
	fmt.Println("errors.As lo reconoce?", errors.As(err, &destino))
	fmt.Println("y destino vale:       ", destino)
	fmt.Println()
	fmt.Println("O sea: As dice que SI es un *MiError, y ese *MiError es nil.")
	fmt.Println("Si ahora lees destino.Codigo, panic:")
	fmt.Println(destino.Codigo)
}
