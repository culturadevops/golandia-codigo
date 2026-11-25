// coste: cuánto cuesta una goroutine. Arranca N goroutines que se quedan esperando en un canal y mide
// el tiempo de arranque y la memoria de pila que reservan.
package main

import (
	"fmt"
	"os"
	"runtime"
	"strconv"
	"sync"
	"time"
)

func main() {
	n, _ := strconv.Atoi(os.Args[1])
	var antes, despues runtime.MemStats
	runtime.GC()
	runtime.ReadMemStats(&antes)
	espera := make(chan struct{})
	var listas sync.WaitGroup
	listas.Add(n)
	t0 := time.Now()
	for i := 0; i < n; i++ {
		go func() {
			listas.Done()
			<-espera
		}()
	}
	listas.Wait()
	arranque := time.Since(t0)
	runtime.ReadMemStats(&despues)
	pila := despues.StackInuse - antes.StackInuse
	fmt.Printf("%7d goroutines · arrancadas en %v · vivas: %d · pila: %.1f MB (%.1f KB cada una)\n",
		n, arranque.Round(time.Millisecond), runtime.NumGoroutine()-1, float64(pila)/1e6, float64(pila)/float64(n)/1024)
	close(espera)
}
