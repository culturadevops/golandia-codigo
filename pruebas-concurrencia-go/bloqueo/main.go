// bloqueo: un envío a un canal sin búfer que nadie va a leer.
package main

import "fmt"

func main() {
	ch := make(chan int)
	fmt.Println("enviando 42 a un canal sin búfer…")
	ch <- 42
	fmt.Println("enviado:", <-ch)
}
