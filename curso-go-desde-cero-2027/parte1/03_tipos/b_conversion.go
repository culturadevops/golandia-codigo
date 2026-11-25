package main

import "fmt"

func main() {
	var tamano int64 = 4096
	bloques := 3
	fmt.Println("total:", tamano*int64(bloques))

	precio := 19.99
	fmt.Println("int(19.99) =", int(precio))

	var nivel int8 = 127
	nivel++
	fmt.Println("int8 127 + 1 =", nivel)

	var grande int64 = 300
	fmt.Println("uint8(300) =", uint8(grande))
}
