package main

import "fmt"

func main() {
	var m map[string]int
	fmt.Println("leer de un mapa nil va bien:", m["x"])
	fmt.Println("ahora escribo...")
	m["x"] = 1 // <-- aqui
}
