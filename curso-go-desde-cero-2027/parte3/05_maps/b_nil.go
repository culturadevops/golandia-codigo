package main

import "fmt"

func main() {
	var visitas map[string]int
	fmt.Println("leer de un map nil:", visitas["inicio"], "len", len(visitas))
	visitas["inicio"]++
	fmt.Println("esto ya no sale")
}
