// Buscar un usuario que puede no existir, y usarlo sin comprobar.
package main

import "fmt"

type Usuario struct{ Nombre string }

var usuarios = map[string]*Usuario{"ana": {Nombre: "Ana"}}

func buscar(id string) *Usuario { return usuarios[id] }

func main() {
	u := buscar("luis")
	fmt.Println("hola,", u.Nombre)
}
