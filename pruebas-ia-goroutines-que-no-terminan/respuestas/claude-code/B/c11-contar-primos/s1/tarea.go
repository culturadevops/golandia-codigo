package tarea

import "sync"

// esPrimo indica si n es un número primo.
func esPrimo(n int) bool {
	if n < 2 {
		return false
	}
	if n%2 == 0 {
		return n == 2
	}
	if n%3 == 0 {
		return n == 3
	}
	for i := 5; i <= n/i; i += 6 {
		if n%i == 0 || n%(i+2) == 0 {
			return false
		}
	}
	return true
}

// ContarPrimos cuenta cuántos números de nums son primos, repartiendo
// el trabajo entre workers goroutines. Si workers <= 0 se usa 1.
// Todas las goroutines lanzadas terminan antes de que la función retorne.
func ContarPrimos(nums []int, workers int) int {
	if len(nums) == 0 {
		return 0
	}
	if workers <= 0 {
		workers = 1
	}
	if workers > len(nums) {
		workers = len(nums)
	}

	parciales := make([]int, workers)
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
		go func(id int, trozo []int) {
			defer wg.Done()
			c := 0
			for _, n := range trozo {
				if esPrimo(n) {
					c++
				}
			}
			parciales[id] = c
		}(w, nums[inicio:fin])
	}
	wg.Wait()

	total := 0
	for _, c := range parciales {
		total += c
	}
	return total
}
