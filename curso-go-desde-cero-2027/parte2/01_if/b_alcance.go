package main

import (
	"fmt"
	"strconv"
)

func main() {
	if n, err := strconv.Atoi("42"); err == nil {
		fmt.Println("dentro:", n)
	}
	fmt.Println("fuera:", n)
}
