```go
package tarea

import (
	"sync"
)

func ContarPrimos(nums []int, workers int) int {
	var mu sync.Mutex
	count := 0

	var wg sync.WaitGroup
	for i := 0; i < workers; i++ {
		wg.Add(1)
		go func(start, end int) {
			defer wg.Done()
			for _, num := range nums[start:end] {
				if esPrimo(num) {
					mu.Lock()
					count++
					mu.Unlock()
				}
			}
		}(i*len(nums)/workers, (i+1)*len(nums)/workers)
	}

	wg.Wait()
	return count
}

func esPrimo(num int) bool {
	if num <= 1 {
		return false
	}
	for i := 2; i*i <= num; i++ {
		if num%i == 0 {
			return false
		}
	}
	return true
}
```