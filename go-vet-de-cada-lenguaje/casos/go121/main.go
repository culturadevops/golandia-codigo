package main

import (
	"fmt"
	"sort"
	"sync"
)

// 4 · La variable de bucle capturada por una goroutine. Con «go 1.21» en go.mod rige la semántica vieja:
// todas las goroutines comparten la misma variable.
func main() {
	urls := []string{"a.com", "b.com", "c.com"}
	var mu sync.Mutex
	var vistas []string
	var wg sync.WaitGroup
	for _, u := range urls {
		wg.Add(1)
		go func() {
			defer wg.Done()
			mu.Lock()
			vistas = append(vistas, u)
			mu.Unlock()
		}()
	}
	wg.Wait()
	sort.Strings(vistas)
	fmt.Println(vistas)
}
