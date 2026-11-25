// Leccion 6 — No hay try/except: el error es un valor que te devuelven.
package main

import (
	"errors"
	"fmt"
	"os"
	"strconv"
)

var ErrPuertoInvalido = errors.New("puerto invalido")

func leerPuerto(s string) (int, error) {
	p, err := strconv.Atoi(s)
	if err != nil {
		return 0, fmt.Errorf("leyendo %q: %w", s, ErrPuertoInvalido)
	}
	return p, nil
}

func main() {
	fmt.Println("== 6.1 · El error viene en el return ==")
	if p, err := leerPuerto("443"); err == nil {
		fmt.Printf("  leerPuerto(\"443\")  -> %d\n", p)
	}
	_, err := leerPuerto("cuatro")
	fmt.Printf("  leerPuerto(\"cuatro\") -> error: %v\n", err)

	fmt.Println()
	fmt.Println("== 6.2 · Envolver y preguntar (el `from e` de Python) ==")
	fmt.Printf("  errors.Is(err, ErrPuertoInvalido) -> %v\n", errors.Is(err, ErrPuertoInvalido))
	fmt.Println("  %w envuelve el error original: se puede preguntar por el.")

	fmt.Println()
	fmt.Println("== 6.3 · LA TRAMPA: el error que tiras a la basura ==")
	f, _ := os.Open("/no/existe/de/verdad.conf")
	fmt.Printf("  f, _ := os.Open(...)  ->  f = %v\n", f)
	fmt.Println("  El _ descarta el error SIN avisar. Tu programa sigue con un")
	fmt.Println("  fichero nil y revienta mas tarde, lejos de la causa.")
	fmt.Println("  En Python la excepcion te paraba AQUI. En Go, solo si la miras.")
}
