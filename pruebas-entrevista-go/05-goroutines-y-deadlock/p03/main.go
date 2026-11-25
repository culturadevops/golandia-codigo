package main

import (
	"fmt"
	"sync"
	"sync/atomic"
)

func main() {
	// EL BUG: el Add(1) esta DENTRO de la goroutine.
	var wg sync.WaitGroup
	var hechas atomic.Int64
	for i := 0; i < 1000; i++ {
		go func() {
			wg.Add(1) // <-- tarde: Wait() puede haber vuelto ya
			defer wg.Done()
			hechas.Add(1)
		}()
	}
	wg.Wait()
	fmt.Println("con Add DENTRO -> esperaba 1000, hechas:", hechas.Load())

	// BIEN: Add antes de lanzar.
	var wg2 sync.WaitGroup
	var hechas2 atomic.Int64
	for i := 0; i < 1000; i++ {
		wg2.Add(1)
		go func() {
			defer wg2.Done()
			hechas2.Add(1)
		}()
	}
	wg2.Wait()
	fmt.Println("con Add FUERA   -> esperaba 1000, hechas:", hechas2.Load())
}
