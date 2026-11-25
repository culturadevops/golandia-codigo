package main

import "fmt"

type Punto struct {
	X, Y int
}

func main() {
	p := Punto{3}
	fmt.Println(p)
}
