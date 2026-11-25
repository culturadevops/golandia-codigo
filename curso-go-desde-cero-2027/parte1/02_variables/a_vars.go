package main

import "fmt"

func main() {
	var nombre string = "Ana"
	var edad int
	vidas := 3
	precio := 19.99
	var activo bool
	var x, y = 1, "dos"

	fmt.Println(nombre, edad, vidas, precio, activo, x, y)
	fmt.Printf("%T %T %T %T %T\n", nombre, edad, vidas, precio, activo)
	fmt.Printf("valores cero: %q %d %v %v\n", "", edad, activo, 0.0)
}
