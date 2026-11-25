package main

import (
	"fmt"
	"sort"
	"sync"
)

// 4b · El mismo bucle con «go 1.27» en go.mod: desde Go 1.22 cada vuelta tiene su variable.
// El «u := u» que todavía escriben los asistentes ya no hace falta.
func main() {
	urls := []string{"a.com", "b.com", "c.com"}
	var mu sync.Mutex
	var vistas []string
	var wg sync.WaitGroup
	for _, u := range urls {
		u := u
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
