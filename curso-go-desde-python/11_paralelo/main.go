// Leccion 11 — Treinta chequeos: uno detras de otro, todos a la vez, y con limite.
package main

import (
	"fmt"
	"sync"
	"time"
)

const N = 30
const LATENCIA = 200 * time.Millisecond // lo que tarda en contestar cada servicio

func chequear(i int) bool { time.Sleep(LATENCIA); return i%7 != 0 }

func main() {
	fmt.Printf("== %d servicios, cada uno tarda %v en contestar ==\n\n", N, LATENCIA)

	t0 := time.Now()
	for i := 0; i < N; i++ {
		chequear(i)
	}
	sec := time.Since(t0)
	fmt.Printf("  uno detras de otro        %6d ms\n", sec.Milliseconds())

	t0 = time.Now()
	var wg sync.WaitGroup
	for i := 0; i < N; i++ {
		wg.Add(1)
		go func() { defer wg.Done(); chequear(i) }()
	}
	wg.Wait()
	par := time.Since(t0)
	fmt.Printf("  todos a la vez            %6d ms\n", par.Milliseconds())

	t0 = time.Now()
	limite := make(chan struct{}, 5) // como mucho 5 a la vez
	for i := 0; i < N; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			limite <- struct{}{}
			chequear(i)
			<-limite
		}()
	}
	wg.Wait()
	lim := time.Since(t0)
	fmt.Printf("  como mucho 5 a la vez     %6d ms\n", lim.Milliseconds())

	fmt.Println()
	fmt.Printf("  todos a la vez es %.0f veces mas rapido que en fila\n", float64(sec)/float64(par))
	fmt.Println("  El limite existe porque 'todos a la vez' contra 3.000 servicios")
	fmt.Println("  es un ataque, no un chequeo. Un canal con hueco hace de semaforo.")
}
