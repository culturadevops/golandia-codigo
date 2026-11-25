// carrera: 1.000 goroutines suman 1.000 veces cada una a un contador compartido.
// sin protección, con sync.Mutex y con sync/atomic. El resultado correcto es 1.000.000.
package main

import (
	"fmt"
	"os"
	"sync"
	"sync/atomic"
)

func main() {
	modo := os.Args[1]
	var total int
	var total64 atomic.Int64
	var mu sync.Mutex
	var wg sync.WaitGroup
	for g := 0; g < 1000; g++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for i := 0; i < 1000; i++ {
				switch modo {
				case "sin":
					total++
				case "mutex":
					mu.Lock()
					total++
					mu.Unlock()
				case "atomic":
					total64.Add(1)
				}
			}
		}()
	}
	wg.Wait()
	if modo == "atomic" {
		total = int(total64.Load())
	}
	fmt.Printf("  %-6s → total = %d (esperado 1000000)\n", modo, total)
}
