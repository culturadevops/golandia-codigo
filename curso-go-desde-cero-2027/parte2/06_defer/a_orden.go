package main

import "fmt"

func main() {
	fmt.Println("abro el archivo")
	defer fmt.Println("defer 1: cierro el archivo")
	defer fmt.Println("defer 2: libero el candado")
	defer fmt.Println("defer 3: anoto en el log")
	fmt.Println("trabajo...")
}
