//go:build !malo

package pedidos

import "sync"

// TotalCarrito pide el precio de cada producto en paralelo y los suma.
// wg.Go (Go 1.25) hace el Add antes de lanzar la goroutine y el Done al terminar.
func TotalCarrito(ids []int, precio func(int) int) int {
	var wg sync.WaitGroup
	var mu sync.Mutex
	total := 0
	for _, id := range ids {
		wg.Go(func() {
			p := precio(id)
			mu.Lock()
			total += p
			mu.Unlock()
		})
	}
	wg.Wait()
	mu.Lock()
	defer mu.Unlock()
	return total
}
