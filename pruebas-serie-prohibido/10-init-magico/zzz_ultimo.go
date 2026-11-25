package main

import "fmt"

// Se llama zzz para comprobar si el orden es ALFABETICO por nombre de fichero.
// Si lo es, este init() tiene que correr DESPUES del de main.go.
func init() { fmt.Println("  3b. init() de zzz_ultimo.go") }
