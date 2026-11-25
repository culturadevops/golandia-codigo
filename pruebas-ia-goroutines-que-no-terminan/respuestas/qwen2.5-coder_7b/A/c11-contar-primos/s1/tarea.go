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
	var wg sync.WaitGroup
	count := 0
	ch := make(chan int, len(nums))

	for i := 0; i < workers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for n := range ch {
				if esPrimo(n) {
					count++
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
