package main

import "fmt"

func main() {
	ch := make(chan int) // SIN buffer

	ch <- 1 // envio... y nadie esta esperando al otro lado

	fmt.Println("esto no se imprime nunca:", <-ch)
}
