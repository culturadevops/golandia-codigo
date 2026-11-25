// Esperar a la red: 50 peticiones a servidores que tardan 50-300 ms, con 50 goroutines.
package main

import (
	"fmt"
	"net/http"
	"sync"
	"time"
)

func main() {
	t := time.Now()
	var ok int
	var mu sync.Mutex
	var wg sync.WaitGroup
	for p := 9001; p <= 9050; p++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			r, err := http.Get(fmt.Sprintf("http://127.0.0.1:%d/salud", p))
			if err == nil && r.StatusCode == 200 {
				mu.Lock()
				ok++
				mu.Unlock()
				r.Body.Close()
			}
		}()
	}
	wg.Wait()
	fmt.Printf("50 goroutines  %d bien · %.2f s\n", ok, time.Since(t).Seconds())
}
