// Los mismos tres errores de C, escritos en Go.
package main

import (
	"fmt"
	"os"
)

func nuevo(v int) *int {
	x := v
	return &x // Go mueve x al heap: es correcto
}

func main() {
	a, b := nuevo(42), nuevo(7)
	fmt.Printf("a=%d b=%d\n", *a, *b)
	nombre := make([]byte, 8)
	texto := "invitado-xx"
	defer func() {
		if r := recover(); r != nil {
			fmt.Println("panic:", r)
			os.Exit(2)
		}
	}()
	for i := 0; i < len(texto); i++ {
		nombre[i] = texto[i]
	}
	fmt.Println(string(nombre))
}
