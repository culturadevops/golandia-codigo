// Cuenta primos hasta N, repartido en 8 trozos: en fila y con 8 goroutines.
package main

import (
	"fmt"
	"sync"
	"time"
)

const N, trozos = 2_000_000, 8

func esPrimo(n int) bool {
	if n < 2 {
		return false
	}
	for i := 2; i*i <= n; i++ {
		if n%i == 0 {
			return false
		}
	}
	return true
}

func contar(a, b int) int {
	c := 0
	for n := a; n < b; n++ {
		if esPrimo(n) {
			c++
		}
	}
	return c
}

func medir(nombre string, f func() int) {
	t := time.Now()
	total := f()
	fmt.Printf("%-14s %d primos · %.2f s\n", nombre, total, time.Since(t).Seconds())
}

func main() {
	paso := N / trozos
	medir("en fila", func() int {
		total := 0
		for i := 0; i < trozos; i++ {
			total += contar(i*paso, (i+1)*paso)
		}
		return total
	})
	medir("8 goroutines", func() int {
		res := make([]int, trozos)
		var wg sync.WaitGroup
		for i := 0; i < trozos; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				res[i] = contar(i*paso, (i+1)*paso)
			}()
		}
		wg.Wait()
		total := 0
		for _, r := range res {
			total += r
		}
		return total
	})
}
