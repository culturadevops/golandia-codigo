// Go 1.27: paquete uuid en la librería estándar.
package main

import (
	"fmt"
	"slices"
	"uuid"
)

func main() {
	a := uuid.New()
	fmt.Println("New():  ", a, "(versión", a.String()[14:15]+")")

	v7 := make([]uuid.UUID, 5)
	for i := range v7 {
		v7[i] = uuid.NewV7()
	}
	for _, u := range v7 {
		fmt.Println("NewV7():", u)
	}
	fmt.Println("¿v7 ya ordenados?", slices.IsSortedFunc(v7, uuid.UUID.Compare))

	p, err := uuid.Parse("{F81D4FAE-7DEC-11D0-A765-00A0C91E6BF6}")
	fmt.Println("Parse con llaves y mayúsculas:", p, err)

	_, err = uuid.Parse("no-es-un-uuid")
	fmt.Println("Parse inválido:", err)

	fmt.Println("¿comparables con ==?", p == uuid.MustParse("f81d4fae7dec11d0a76500a0c91e6bf6"))
}
