package main

import (
	"errors"
	"fmt"
)

type ErrHTTP struct {
	Codigo int
	URL    string
}

func (e *ErrHTTP) Error() string { return fmt.Sprintf("%s devolvio %d", e.URL, e.Codigo) }

func llamar() error {
	return fmt.Errorf("sincronizando: %w", &ErrHTTP{Codigo: 429, URL: "/api/pedidos"})
}

func main() {
	err := llamar()
	fmt.Println("el error:", err)

	// Is compara IDENTIDAD. As extrae el TIPO para poder leerle los campos.
	var e *ErrHTTP
	if errors.As(err, &e) {
		fmt.Println("errors.As lo saco. Codigo =", e.Codigo, "URL =", e.URL)
		if e.Codigo == 429 {
			fmt.Println("-> puedo decidir: esperar y reintentar")
		}
	}
	fmt.Println()
	fmt.Println("Esa es la diferencia en una frase:")
	fmt.Println("  Is  = es ESTE error?           (necesitas comparar)")
	fmt.Println("  As  = es de ESTE TIPO, damelo. (necesitas los datos)")
}
