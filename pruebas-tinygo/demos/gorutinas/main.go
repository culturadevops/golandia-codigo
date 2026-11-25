// El cálculo de /suma (200.000 vueltas), 400 veces: en serie y con 2 goroutines, tras una pasada de calentamiento. Mismo archivo para go build y tinygo build.
package main

import (
	"fmt"
	"os"
	"runtime"
	"strconv"
	"sync"
	"time"
)

func suma(n int) uint64 {
	var s uint64
	for i := 1; i <= n; i++ {
		s += uint64(i) * uint64(i) % 1000003
	}
	return s
}

var sumidero uint64

func main() {
	// n se lee al ejecutar (variable N): con una constante, LLVM (TinyGo) calcula suma(n) al compilar y mide 0 ms.
	n, err := strconv.Atoi(os.Getenv("N"))
	if err != nil {
		n = 200_000
	}
	const veces = 400
	fmt.Println("go", runtime.Version(), "· GOMAXPROCS:", runtime.GOMAXPROCS(0))
	for i := 0; i < veces; i++ { // calentamiento, no se mide
		sumidero += suma(n)
	}
	t := time.Now()
	for i := 0; i < veces; i++ {
		sumidero += suma(n)
	}
	fmt.Printf("en serie:        %d ms\n", time.Since(t).Milliseconds())
	t = time.Now()
	var wg sync.WaitGroup
	var mu sync.Mutex
	for g := 0; g < 2; g++ {
		wg.Go(func() {
			var local uint64
			for i := 0; i < veces/2; i++ {
				local += suma(n)
			}
			mu.Lock()
			sumidero += local
			mu.Unlock()
		})
	}
	wg.Wait()
	fmt.Printf("2 goroutines:    %d ms\n", time.Since(t).Milliseconds())
	// Se imprime el resultado: si nadie lo lee, LLVM (TinyGo) borra el cálculo entero y mide 0 ms.
	fmt.Println("comprobación:", sumidero)
}
