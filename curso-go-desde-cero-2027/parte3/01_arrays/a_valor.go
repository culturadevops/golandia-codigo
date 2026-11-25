package main

import "fmt"

func main() {
	a := [3]int{1, 2, 3}
	b := a
	b[0] = 99
	fmt.Println("a =", a, " b =", b)
	fmt.Println("a == [3]int{1, 2, 3}:", a == [3]int{1, 2, 3})

	s := []int{1, 2, 3}
	t := s
	t[0] = 99
	fmt.Println("s =", s, " t =", t)
}
