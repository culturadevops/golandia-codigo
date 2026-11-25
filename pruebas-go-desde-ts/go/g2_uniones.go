//go:build ignore

// 2 · Aqui NO hay uniones de literales. `type Estado string` no restringe nada:
// acepta cualquier cadena, y el error se descubre en ejecucion o nunca.
package main

import "fmt"

type Estado string

const (
	Activo  Estado = "activo"
	Baja    Estado = "baja"
	Moroso  Estado = "moroso"
)

func describir(e Estado) string {
	return "el cliente esta " + string(e)
}

func main() {
	fmt.Println(describir(Activo))
	fmt.Println(describir("activa")) // compila. Nadie se queja. Nunca.
	fmt.Println(describir("cualquier cosa que se me ocurra"))
}
