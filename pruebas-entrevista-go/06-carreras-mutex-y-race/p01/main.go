package main

import (
	"fmt"
	"sync"
)

func main() {
	contador := 0
	var wg sync.WaitGroup
	for i := 0; i < 2; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < 100000; j++ {
				contador++ // dos goroutines, la misma variable
			}
		}()
	}
	wg.Wait()
	fmt.Println("esperaba 200000, salio:", contador)
}
