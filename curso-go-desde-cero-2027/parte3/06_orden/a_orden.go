package main

import (
	"fmt"
	"maps"
	"slices"
)

func main() {
	m := map[string]int{"uno": 1, "dos": 2, "tres": 3, "cuatro": 4, "cinco": 5}
	for vuelta := 1; vuelta <= 5; vuelta++ {
		fmt.Print("vuelta ", vuelta, ": ")
		for k := range m {
			fmt.Print(k, " ")
		}
		fmt.Println()
	}
	fmt.Println("ordenadas:", slices.Sorted(maps.Keys(m)))
}
