package main

import "fmt"

const Enorme = 1 << 100

func main() {
	fmt.Println(Enorme >> 98)
	fmt.Println(Enorme / 1e30)
	var n int = Enorme
	fmt.Println(n)
}
