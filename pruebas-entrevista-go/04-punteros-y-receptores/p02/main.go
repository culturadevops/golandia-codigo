package main

import "fmt"

type Guardable interface{ Guardar() error }

type Usuario struct{ Nombre string }

// El metodo esta en *Usuario, no en Usuario.
func (u *Usuario) Guardar() error { return nil }

func persistir(g Guardable) { fmt.Println("guardando", g) }

func main() {
	u := Usuario{Nombre: "ana"}
	persistir(u) // <-- pasamos el VALOR, no el puntero
}
