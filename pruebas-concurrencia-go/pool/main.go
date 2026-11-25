// pool: un grupo de trabajadores leyendo de un canal de tareas.
// 1) 200 tareas que ESPERAN 20 ms (como una llamada de red), con 1, 4, 16, 64 y 200 trabajadores;
// 2) 64 tareas que CALCULAN (solo CPU), con 1, 2, 4, 8, 16 y 32 trabajadores.
package main

import (
	"crypto/sha256"
	"fmt"
	"runtime"
	"sync"
	"time"
)

func pool(trabajadores, tareas int, hacer func()) time.Duration {
	ch := make(chan int)
	var wg sync.WaitGroup
	t0 := time.Now()
	for w := 0; w < trabajadores; w++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for range ch {
				hacer()
			}
		}()
	}
	for i := 0; i < tareas; i++ {
		ch <- i
	}
	close(ch)
	wg.Wait()
	return time.Since(t0)
}

func esperar() { time.Sleep(20 * time.Millisecond) }

func calcular() {
	b := []byte("golandia")
	for i := 0; i < 60_000; i++ {
		h := sha256.Sum256(b)
		b = h[:]
	}
}

func main() {
	fmt.Printf("  [GOMAXPROCS=%d · núcleos que ve Go: %d]\n", runtime.GOMAXPROCS(0), runtime.NumCPU())
	fmt.Println("  --- 200 tareas que esperan 20 ms (E/S) ---")
	for _, w := range []int{1, 4, 16, 64, 200} {
		fmt.Printf("  %3d trabajadores: %v\n", w, pool(w, 200, esperar).Round(time.Millisecond))
	}
	fmt.Println("  --- 64 tareas que solo calculan (CPU) ---")
	base := pool(1, 64, calcular)
	fmt.Printf("  %3d trabajadores: %v\n", 1, base.Round(time.Millisecond))
	for _, w := range []int{2, 4, 8, 16, 32} {
		d := pool(w, 64, calcular)
		fmt.Printf("  %3d trabajadores: %v (%.1f× más rápido)\n", w, d.Round(time.Millisecond), float64(base)/float64(d))
	}
}
