// cancelar: 8 trabajadores en bucle; a los 107 ms se cancela el contexto. ¿Cuánto tardan en parar todos?
// (107 y no 100: con 100, múltiplo de 20, todos estaban justo comprobando y paraban al instante.)
// Unos miran ctx.Done() en cada vuelta (select); otros solo cada 20 ms de trabajo.
package main

import (
	"context"
	"fmt"
	"sync"
	"time"
)

func main() {
	for _, cada := range []time.Duration{time.Millisecond, 20 * time.Millisecond} {
		ctx, cancel := context.WithCancel(context.Background())
		var wg sync.WaitGroup
		var mu sync.Mutex
		var ultima time.Time
		vueltas := 0
		for w := 0; w < 8; w++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				for {
					select {
					case <-ctx.Done():
						mu.Lock()
						if t := time.Now(); t.After(ultima) {
							ultima = t
						}
						mu.Unlock()
						return
					default:
					}
					time.Sleep(cada) // una unidad de trabajo
					mu.Lock()
					vueltas++
					mu.Unlock()
				}
			}()
		}
		time.Sleep(107 * time.Millisecond)
		t := time.Now()
		cancel()
		wg.Wait()
		fmt.Printf("  trabajo de %v entre comprobaciones: los 8 pararon %v después de cancel() · %d vueltas hechas\n",
			cada, ultima.Sub(t).Round(100*time.Microsecond), vueltas)
	}
}
