//go:build malo

package pedidos

import "sync"

// TotalCarrito pide el precio de cada producto en paralelo y los suma.
func TotalCarrito(ids []int, precio func(int) int) int {
	var wg sync.WaitGroup
	var mu sync.Mutex
	total := 0
	for _, id := range ids {
		go func() {
			wg.Add(1)
			defer wg.Done()
			p := precio(id)
			mu.Lock()
			total += p
			mu.Unlock()
		}()
	}
	wg.Wait()
	mu.Lock()
	defer mu.Unlock()
	return total
}
