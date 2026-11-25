package main

import (
	"fmt"
	"sync"
)

func main() {
	contador := 0
	var mu sync.Mutex
	var wg sync.WaitGroup
	for range 1000 {
		wg.Go(func() {
			for range 1000 {
				mu.Lock()
				contador++
				mu.Unlock()
			}
		})
	}
	wg.Wait()
	fmt.Println("contador =", contador, "(esperaba 1000000)")
}
