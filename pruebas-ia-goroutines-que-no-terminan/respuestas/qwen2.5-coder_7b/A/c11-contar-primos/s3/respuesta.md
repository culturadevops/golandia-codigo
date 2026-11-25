```go
package tarea

import (
	"sync"
)

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

func contarPrimos(nums []int, workers int) int {
	var wg sync.WaitGroup
	ch := make(chan int)
	result := 0

	for i := 0; i < workers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for num := range ch {
				if esPrimo(num) {
					result++
				}
			}
		}()
	}

	for _, num := range nums {
		ch <- num
	}
	close(ch)

	wg.Wait()
	return result
}
```