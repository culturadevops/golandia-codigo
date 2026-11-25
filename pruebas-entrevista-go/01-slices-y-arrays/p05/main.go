package main

import "fmt"

func main() {
	// Como CRECE la capacidad: solo se imprime cuando cambia.
	var s []int
	prev := -1
	for i := 0; i < 2000; i++ {
		if cap(s) != prev {
			fmt.Printf("len %-5d cap %d\n", len(s), cap(s))
			prev = cap(s)
		}
		s = append(s, i)
	}
	fmt.Println("final -> len", len(s), "cap", cap(s))
}
