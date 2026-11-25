package tarea

import (
	"sync"
)

func esPrimo(n int) bool {
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

func contarPrimos(nums []int, workers int) int {
	var wg sync.WaitGroup
	ch := make(chan int, workers)
	count := 0

	for i := 0; i < workers; i++ {
		wg.Add(1)
		go func(start, end int) {
			defer wg.Done()
			for _, num := range nums[start:end] {
				if esPrimo(num) {
					ch <- 1
				} else {
					ch <- 0
				}
			}
		}(i*len(nums)/workers, (i+1)*len(nums)/workers)
	}

	go func() {
		wg.Wait()
		close(ch)
	}()

	for num := range ch {
		count += num
	}

	return count
}
