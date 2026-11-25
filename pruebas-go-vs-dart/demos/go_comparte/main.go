// Un contador global, sumado desde 2 goroutines (1.000 veces cada una), sin y con mutex.
package main

import (
	"fmt"
	"sync"
)

var contador int

func main() {
	var wg sync.WaitGroup
	for g := 0; g < 2; g++ {
		wg.Go(func() {
			for i := 0; i < 1000; i++ {
				contador++
			}
		})
	}
	wg.Wait()
	fmt.Println("el contador compartido:", contador)
}
