package main

import (
	"fmt"
	"sync"
)

// Repite la carrera N veces y CUENTA cuantas dan el numero correcto.
// Una anecdota ("a mi me salio bien una vez") no es un dato. Esto si.
func unaCorrida() int {
	c := 0
	var wg sync.WaitGroup
	for i := 0; i < 2; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < 100000; j++ {
				c++
			}
		}()
	}
	wg.Wait()
	return c
}

func main() {
	const N = 500
	exactos, min, max := 0, 1<<62, 0
	for i := 0; i < N; i++ {
		v := unaCorrida()
		if v == 200000 {
			exactos++
		}
		if v < min {
			min = v
		}
		if v > max {
			max = v
		}
	}
	fmt.Printf("corridas: %d\n", N)
	fmt.Printf("dieron EXACTAMENTE 200000: %d\n", exactos)
	fmt.Printf("porcentaje que parecia correcto: %.1f%%\n", float64(exactos)*100/float64(N))
	fmt.Printf("peor caso: %d   mejor caso: %d\n", min, max)
}
