package main

import "fmt"

func main() {
	ch := make(chan int, 2)
	ch <- 1
	close(ch)

	// Recibir de un canal cerrado va bien: da el valor cero y ok=false.
	v, ok := <-ch
	fmt.Println("primero: ", v, ok)
	v, ok = <-ch
	fmt.Println("ya vacio:", v, ok, " <-- cero y false, NO bloquea")

	// Enviar a un canal cerrado, en cambio:
	fmt.Println("ahora envio a un canal cerrado...")
	ch <- 3
}
