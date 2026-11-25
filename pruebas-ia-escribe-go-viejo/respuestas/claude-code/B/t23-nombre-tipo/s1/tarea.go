package tarea

import "reflect"

// NombreTipo devuelve el nombre del tipo T, incluidas las interfaces como error.
// reflect.TypeFor obtiene el tipo estático de T, así que no falla con interfaces,
// a diferencia de reflect.TypeOf(valorCero), que devuelve nil para ellas.
func NombreTipo[T any]() string {
	return reflect.TypeFor[T]().String()
}
