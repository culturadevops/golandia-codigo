```go
package tarea

import (
	"sync"
)

func ContarPrimos(nums []int, workers int) int {
	var mu sync.Mutex
	var count int

	isPrime := func(n int) bool {
		if n <= 1 {
			return false
		}
		for i := 2; i*i <= n; i++ {
			if n%i == 0 {
				return false
			}
		}
		return true
	}

	ch := make(chan int)
	wg := sync.WaitGroup{}

	for i := 0; i < workers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for n := range ch {
				if isPrime(n) {
					mu.Lock()
					count++
					mu.Unlock()
				}
			}
		}()
	}

	for _, n := range nums {
		ch <- n
	}
	close(ch)

	wg.Wait()

	return count
}
```