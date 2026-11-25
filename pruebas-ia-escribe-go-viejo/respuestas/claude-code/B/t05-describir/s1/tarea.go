package tarea

import "fmt"

// Describir devuelve el tipo de cada valor recibido, por ejemplo "int" o "string".
func Describir(valores ...any) []string {
	tipos := make([]string, 0, len(valores))
	for _, v := range valores {
		tipos = append(tipos, fmt.Sprintf("%T", v))
	}
	return tipos
}
