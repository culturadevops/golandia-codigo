package main

import "sync"

func main() {
	m := map[int]int{}
	var wg sync.WaitGroup
	for i := 0; i < 8; i++ {
		wg.Add(1)
		go func(n int) {
			defer wg.Done()
			for j := 0; j < 10000; j++ {
				m[n*10000+j] = j // dos goroutines escribiendo el mismo mapa
			}
		}(i)
	}
	wg.Wait()
}
