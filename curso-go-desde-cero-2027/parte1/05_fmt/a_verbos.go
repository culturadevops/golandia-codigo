package main

import "fmt"

type Punto struct {
	X, Y int
}

func main() {
	p := Punto{3, 4}
	fmt.Printf("%v | %+v | %#v | %T\n", p, p, p, p)
	fmt.Printf("%d | %5d | %05d | %x | %b\n", 42, 42, 42, 255, 5)
	fmt.Printf("%f | %.2f | %8.3f | %e\n", 3.14159, 3.14159, 3.14159, 123456.789)
	fmt.Printf("%s | %q | %t\n", "hola", "hola", true)
	s := fmt.Sprintf("%s tiene %d años", "Ana", 30)
	fmt.Println(s)
}
