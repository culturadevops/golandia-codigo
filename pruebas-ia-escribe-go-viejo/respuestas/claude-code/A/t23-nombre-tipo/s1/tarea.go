package tarea

import "reflect"

// NombreTipo devuelve el nombre del tipo T, incluso cuando T es una interfaz
// (por ejemplo, error). Para tipos sin nombre (como []int o map[string]int)
// devuelve su representación textual.
func NombreTipo[T any]() string {
	t := reflect.TypeOf((*T)(nil)).Elem()
	if nombre := t.Name(); nombre != "" {
		return nombre
	}
	return t.String()
}
