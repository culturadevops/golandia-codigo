package main

import "fmt"

func dividir(a, b int) (int, int) {
	return a / b, a % b
}

func main() {
	c := dividir(17, 5)
	fmt.Println(c)
}
