package main

import (
	"fmt"
	"sync"
)

func main() {
	contador := 0
	var wg sync.WaitGroup
	for range 1000 {
		wg.Go(func() {
			for range 1000 {
				contador++
			}
		})
	}
	wg.Wait()
	fmt.Println("contador =", contador, "(esperaba 1000000)")
}
