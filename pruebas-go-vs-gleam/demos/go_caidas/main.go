// Tres goroutines trabajan; una entra en pánico a la mitad. ¿Qué les pasa a las otras dos?
package main

import (
	"fmt"
	"time"
)

func trabajar(nombre string, falla bool) {
	time.Sleep(50 * time.Millisecond)
	if falla {
		panic("división entre cero en el pedido 42")
	}
	fmt.Println(nombre + ": terminé")
}

func main() {
	go trabajar("goroutine 1", false)
	go trabajar("goroutine 2", true)
	go trabajar("goroutine 3", false)
	time.Sleep(300 * time.Millisecond)
	fmt.Println("main: sigo vivo")
}
