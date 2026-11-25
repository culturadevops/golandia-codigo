package main

import "fmt"

func main() {
	nivel := 2
	fmt.Println("sin fallthrough:")
	switch nivel {
	case 2:
		fmt.Println("  nivel 2")
	case 1:
		fmt.Println("  nivel 1")
	}
	fmt.Println("con fallthrough:")
	switch nivel {
	case 2:
		fmt.Println("  nivel 2")
		fallthrough
	case 1:
		fmt.Println("  nivel 1")
	case 0:
		fmt.Println("  nivel 0")
	}
}
