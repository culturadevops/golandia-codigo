```go
package tarea

import (
	"sync"
)

func ContarPrimos(nums []int, workers int) int {
	var wg sync.WaitGroup
	ch := make(chan int)
	result := 0

	// Función para verificar si un número es primo
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

	// Goroutine para contar primos
	go func() {
		for n := range ch {
			if isPrime(n) {
				result++
			}
		}
	}()

	// Distribuir trabajo entre workers goroutines
	for i := 0; i < workers; i++ {
		wg.Add(1)
		go func(start, end int) {
			defer wg.Done()
			for n := start; n < end; n++ {
				ch <- nums[n]
			}
		}(i*len(nums)/workers, (i+1)*len(nums)/workers)
	}

	// Cerrar el canal después de enviar todos los números
	close(ch)

	// Esperar a que todas las goroutines terminen
	wg.Wait()

	return result
}
```