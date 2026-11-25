package main

import "fmt"

func main() {
	var s []int
	fmt.Println("nil:", s == nil, "len", len(s), "cap", cap(s))
	ultimaCap := cap(s)
	for i := 0; i < 2000; i++ {
		s = append(s, i)
		if cap(s) != ultimaCap {
			fmt.Printf("len %4d -> cap %4d\n", len(s), cap(s))
			ultimaCap = cap(s)
		}
	}
}
