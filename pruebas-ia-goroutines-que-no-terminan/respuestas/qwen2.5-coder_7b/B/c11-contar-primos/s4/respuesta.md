```go
package tarea

import (
	"sync"
)

func esPrimo(n int) bool {
	if n <= 1 {
		return false
	}
	if n <= 3 {
		return true
	}
	if n%2 == 0 || n%3 == 0 {
		return false
	}
	for i := 5; i*i <= n; i += 6 {
		if n%i == 0 || n%(i+2) == 0 {
			return false
		}
	}
	return true
}

func ContarPrimos(nums []int, workers int) int {
	var mu sync.Mutex
	count := 0

	sem := make(chan struct{}, workers)

	for _, num := range nums {
		sem <- struct{}{}
		go func(num int) {
			defer func() { <-sem }()
			if esPrimo(num) {
				mu.Lock()
				count++
				mu.Unlock()
			}
		}(num)
	}

	close(sem)

	return count
}
```