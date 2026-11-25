package main

import "fmt"

func main() {
	var tamano int64 = 4096
	bloques := 3
	total := tamano * bloques
	fmt.Println(total)
}
