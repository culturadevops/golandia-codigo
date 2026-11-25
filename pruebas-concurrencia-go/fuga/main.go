// fuga: una función que lanza una goroutine para buscar algo y se rinde a los 10 ms.
// La goroutine tarda 50 ms y quiere entregar el resultado por un canal que ya nadie lee.
// Se llama 1.000 veces y se cuentan las goroutines vivas.
package main

import (
	"context"
	"fmt"
	"os"
	"runtime"
	"time"
)

func buscarConFuga() {
	res := make(chan int) // sin búfer: el envío espera a un receptor que ya se fue
	go func() {
		time.Sleep(50 * time.Millisecond)
		res <- 42
	}()
	select {
	case <-res:
	case <-time.After(10 * time.Millisecond):
	}
}

func buscarConBufer() {
	res := make(chan int, 1) // un hueco: el envío no espera a nadie
	go func() {
		time.Sleep(50 * time.Millisecond)
		res <- 42
	}()
	select {
	case <-res:
	case <-time.After(10 * time.Millisecond):
	}
}

func buscarConContexto() {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
	defer cancel()
	res := make(chan int)
	go func() {
		select {
		case <-time.After(50 * time.Millisecond):
		case <-ctx.Done():
			return // se entera de que ya no hace falta y termina
		}
		select {
		case res <- 42:
		case <-ctx.Done():
		}
	}()
	select {
	case <-res:
	case <-ctx.Done():
	}
}

func main() {
	f := map[string]func(){"fuga": buscarConFuga, "bufer": buscarConBufer, "contexto": buscarConContexto}[os.Args[1]]
	for i := 0; i < 1000; i++ {
		f()
	}
	fmt.Printf("  %-8s · tras 1.000 llamadas: %d goroutines vivas\n", os.Args[1], runtime.NumGoroutine())
	time.Sleep(200 * time.Millisecond)
	fmt.Printf("  %-8s · 200 ms después:     %d goroutines vivas\n", os.Args[1], runtime.NumGoroutine())
}
