package main

import "fmt"

func main() {
	// ARREGLO A: darle buzon. Cambia el comportamiento: ya no hay cita.
	a := make(chan int, 1)
	a <- 1
	fmt.Println("con buffer:", <-a)

	// ARREGLO B: que haya alguien al otro lado. Es el arreglo BUENO en el 99% de los casos.
	b := make(chan int)
	go func() { b <- 1 }()
	fmt.Println("con goroutine:", <-b)

	fmt.Println()
	fmt.Println("No son intercambiables: el buffer desacopla en el tiempo,")
	fmt.Println("la goroutine mantiene la cita. Elegir mal el arreglo es la trampa.")
}
