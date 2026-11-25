package main

import "fmt"

func main() {
	original := []int{1, 2, 3, 4, 5}
	trozo := original[1:3]

	fmt.Println("original:", original, "len", len(original), "cap", cap(original))
	fmt.Println("trozo:   ", trozo, "len", len(trozo), "cap", cap(trozo))

	trozo = append(trozo, 99)

	fmt.Println("--- despues del append ---")
	fmt.Println("trozo:   ", trozo)
	fmt.Println("original:", original, "  <-- nadie lo toco, y cambio")
}
