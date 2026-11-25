package main

import "fmt"

func main() {
	notas := []int{10, 20, 30, 40, 50}
	primeras := notas[:2]
	fmt.Println("notas:", notas, " primeras:", primeras, "len", len(primeras), "cap", cap(primeras))

	primeras[0] = 99
	fmt.Println("tras primeras[0] = 99 ->", notas)

	primeras = append(primeras, 7)
	fmt.Println("tras append(primeras, 7) ->", notas)
}
