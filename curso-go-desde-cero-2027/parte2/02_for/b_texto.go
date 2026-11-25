package main

import "fmt"

func main() {
	ave := "ñandú"
	fmt.Println("len:", len(ave))
	for i, r := range ave {
		fmt.Printf("%d:%c ", i, r)
	}
	fmt.Println()
	for i := 0; i < len(ave); i++ {
		fmt.Printf("%d ", ave[i])
	}
	fmt.Println()
}
