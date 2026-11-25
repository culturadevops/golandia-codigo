package main

import (
	"fmt"
	"runtime"
	"time"
)

// La fuga: la goroutine se queda esperando para siempre en un canal que nadie leera.
func fuga() {
	ch := make(chan int)
	go func() {
		ch <- 42 // nadie va a recibir: esta goroutine no muere jamas
	}()
	// nos vamos sin leer
}

func main() {
	fmt.Println("goroutines al empezar:", runtime.NumGoroutine())
	for i := 0; i < 100; i++ {
		fuga()
	}
	time.Sleep(100 * time.Millisecond)
	fmt.Println("goroutines tras 100 llamadas:", runtime.NumGoroutine())
	fmt.Println()
	fmt.Println("El programa acaba sin error y sin aviso. Nadie te dice nada.")
	fmt.Println("En un servidor, esto crece hasta que el proceso se muere por memoria.")
}
