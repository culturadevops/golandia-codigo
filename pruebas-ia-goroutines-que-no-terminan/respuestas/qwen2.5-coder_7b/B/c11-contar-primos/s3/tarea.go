package tarea

import (
	"sync"
)

func ContarPrimos(nums []int, workers int) int {
	var mu sync.Mutex
	count := 0

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

	// Canal para enviar números a los workers
	numChan := make(chan int, len(nums))
	// Canal para señalizar el fin del trabajo
	done := make(chan struct{})

	// Iniciar los workers
	for i := 0; i < workers; i++ {
		go func() {
			for n := range numChan {
				if isPrime(n) {
					mu.Lock()
					count++
					mu.Unlock()
				}
			}
			done <- struct{}{}
		}()
	}

	// Enviar números al canal
	for _, n := range nums {
		numChan <- n
	}
	close(numChan)

	// Esperar a que todos los workers terminen
	for i := 0; i < workers; i++ {
		<-done
	}

	return count
}
