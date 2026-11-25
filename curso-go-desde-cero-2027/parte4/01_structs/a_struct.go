package main

import "fmt"

type Usuario struct {
	Nombre string
	Edad   int
	Activo bool
}

func main() {
	ana := Usuario{Nombre: "Ana", Edad: 31, Activo: true}
	var vacio Usuario
	beto := Usuario{Nombre: "Beto"}
	fmt.Println(ana.Nombre, ana.Edad, ana.Activo)
	fmt.Printf("vacío: %+v\n", vacio)
	fmt.Printf("beto:  %+v\n", beto)
	ana.Edad++
	fmt.Println("ana cumple años:", ana.Edad)
	otra := Usuario{Nombre: "Ana", Edad: 32, Activo: true}
	fmt.Println("¿ana == otra?", ana == otra)
}
