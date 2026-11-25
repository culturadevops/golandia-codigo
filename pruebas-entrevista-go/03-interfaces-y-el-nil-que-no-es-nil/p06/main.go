package main

import (
	"fmt"
	"reflect"
)

type MiError struct{ Codigo int }

func (e *MiError) Error() string { return "fallo" }

// Lo que la gente escribe cuando descubre el problema. Funciona... y es la respuesta MALA.
func esRealmenteNil(v any) bool {
	if v == nil {
		return true
	}
	rv := reflect.ValueOf(v)
	switch rv.Kind() {
	case reflect.Ptr, reflect.Map, reflect.Slice, reflect.Chan, reflect.Func:
		return rv.IsNil()
	}
	return false
}

func main() {
	var p *MiError
	var iface error = p
	fmt.Println("iface == nil?          ", iface == nil)
	fmt.Println("esRealmenteNil(iface)? ", esRealmenteNil(iface))
	fmt.Println()
	fmt.Println("Funciona. Pero ahora cada comprobacion de error de tu programa")
	fmt.Println("pasa por reflect, y el bug sigue estando en la funcion que lo devuelve.")
}
