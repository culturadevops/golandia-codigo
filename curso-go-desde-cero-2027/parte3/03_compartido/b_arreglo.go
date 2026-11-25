package main

import (
	"fmt"
	"slices"
)

func main() {
	notas := []int{10, 20, 30, 40, 50}
	a := notas[:2:2]
	a = append(a, 7)
	fmt.Println("con notas[:2:2]:", notas, a)

	notas2 := []int{10, 20, 30, 40, 50}
	b := slices.Clone(notas2[:2])
	b = append(b, 7)
	fmt.Println("con slices.Clone:", notas2, b)
}
