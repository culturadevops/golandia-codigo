package main

import "fmt"

func main() {
	contador := 1
	defer fmt.Println("defer con argumento:", contador)
	defer func() { fmt.Println("defer con closure:", contador) }()
	contador = 100
	fmt.Println("al final de main, contador =", contador)
}
