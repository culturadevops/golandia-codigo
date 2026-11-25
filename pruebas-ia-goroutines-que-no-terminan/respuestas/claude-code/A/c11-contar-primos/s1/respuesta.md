```go
package tarea

import "sync"

func esPrimo(n int) bool {
	if n < 2 {
		return false
	}
	if n < 4 {
		return true
	}
	if n%2 == 0 || n%3 == 0 {
		return false
	}
	for i := 5; i <= n/i; i += 6 {
		if n%i == 0 || n%(i+2) == 0 {
			return false
		}
	}
	return true
}

// ContarPrimos cuenta cuántos números de nums son primos, repartiendo el
// trabajo entre workers goroutines.
func ContarPrimos(nums []int, workers int) int {
	if len(nums) == 0 {
		return 0
	}
	if workers < 1 {
		workers = 1
	}
	if workers > len(nums) {
		workers = len(nums)
	}

	resultados := make([]int, workers)
	tam := (len(nums) + workers - 1) / workers

	var wg sync.WaitGroup
	for w := 0; w < workers; w++ {
		inicio := w * tam
		if inicio >= len(nums) {
			break
		}
		fin := inicio + tam
		if fin > len(nums) {
			fin = len(nums)
		}
		wg.Add(1)
		go func(w int, parte []int) {
			defer wg.Done()
			c := 0
			for _, n := range parte {
				if esPrimo(n) {
					c++
				}
			}
			resultados[w] = c
		}(w, nums[inicio:fin])
	}
	wg.Wait()

	total := 0
	for _, c := range resultados {
		total += c
	}
	return total
}
```