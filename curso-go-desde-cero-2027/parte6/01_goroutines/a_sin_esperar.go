package main

import "fmt"

func saludar(n int) {
	fmt.Println("hola desde la goroutine", n)
}

func main() {
	for i := 1; i <= 3; i++ {
		go saludar(i)
	}
	fmt.Println("main terminó")
}
