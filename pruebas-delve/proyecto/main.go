package main

import (
	"fmt"
	"sync"
	"time"
)

var mu1, mu2 sync.Mutex

func trabajadorA() {
	mu1.Lock()
	time.Sleep(100 * time.Millisecond)
	mu2.Lock()
	fmt.Println("A termino")
	mu2.Unlock()
	mu1.Unlock()
}

func trabajadorB() {
	mu2.Lock()
	time.Sleep(100 * time.Millisecond)
	mu1.Lock()
	fmt.Println("B termino")
	mu1.Unlock()
	mu2.Unlock()
}

func main() {
	go func() {
		for {
			time.Sleep(1 * time.Second)
			fmt.Println("tick: el programa sigue vivo")
		}
	}()

	go trabajadorA()
	go trabajadorB()

	mensaje := "el deadlock ya deberia estar armado aca"
	time.Sleep(500 * time.Millisecond)
	fmt.Println("punto de inspeccion:", mensaje)
	segundoMensaje := "un paso mas, sin correr todo el programa"
	fmt.Println(segundoMensaje)
	time.Sleep(60 * time.Second)
}
