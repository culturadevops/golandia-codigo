// Go 1.27: la clave de un literal de struct puede ser cualquier selector de campo válido,
// también un campo promovido de un struct embebido.
package main

import "fmt"

type Base struct {
	ID     int
	Creado string
}

type Usuario struct {
	Base
	Nombre string
}

func main() {
	u := Usuario{ID: 7, Creado: "2026-10-29", Nombre: "ana"}
	fmt.Printf("%+v\n", u)
}
