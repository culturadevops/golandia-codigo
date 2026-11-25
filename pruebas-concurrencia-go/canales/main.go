// canales: sin búfer y con búfer. 1) un millón de mensajes de un productor a un consumidor;
// 2) contrapresión: el productor manda cada 1 ms, el consumidor tarda 5 ms; cuánto se bloquea el productor.
package main

import (
	"fmt"
	"time"
)

func transferir(buf, n int) time.Duration {
	ch := make(chan int, buf)
	t0 := time.Now()
	go func() {
		for i := 0; i < n; i++ {
			ch <- i
		}
		close(ch)
	}()
	for range ch {
	}
	return time.Since(t0)
}

func contrapresion(buf int) {
	ch := make(chan int, buf)
	listo := make(chan struct{})
	go func() {
		for range ch {
			time.Sleep(5 * time.Millisecond)
		}
		close(listo)
	}()
	var bloqueado time.Duration
	t0 := time.Now()
	for i := 0; i < 40; i++ {
		a := time.Now()
		ch <- i
		bloqueado += time.Since(a)
		time.Sleep(time.Millisecond)
	}
	close(ch)
	<-listo
	fmt.Printf("  búfer %3d: 40 mensajes en %v · el productor pasó %v bloqueado en el envío\n",
		buf, time.Since(t0).Round(time.Millisecond), bloqueado.Round(time.Millisecond))
}

func main() {
	const n = 1_000_000
	for _, buf := range []int{0, 1, 100, 1000} {
		d := transferir(buf, n)
		fmt.Printf("  búfer %4d: 1.000.000 mensajes en %v (%.0f ns por mensaje)\n", buf, d.Round(time.Millisecond), float64(d.Nanoseconds())/n)
	}
	fmt.Println("  --- contrapresión: el productor manda cada 1 ms, el consumidor tarda 5 ms ---")
	for _, buf := range []int{0, 10, 100} {
		contrapresion(buf)
	}
}
