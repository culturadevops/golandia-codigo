// Leccion 10 — Goroutines: dos fallos que NO dan error.
package main

import (
	"fmt"
	"sync"
	"time"
)

func main() {
	fmt.Println("== 10.1 · LA TRAMPA 1: main no espera a nadie ==")
	hecho := 0
	for i := 0; i < 3; i++ {
		go func() {
			time.Sleep(50 * time.Millisecond)
			hecho++
		}()
	}
	fmt.Printf("  goroutines terminadas al llegar aqui: %d de 3\n", hecho)
	fmt.Println("  Si main acaba, el programa acaba, y lo que estaba a medias se pierde.")

	fmt.Println()
	fmt.Println("== 10.2 · WaitGroup: esperar de verdad ==")
	var wg sync.WaitGroup
	fin := 0
	var mu sync.Mutex
	for i := 0; i < 3; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			time.Sleep(50 * time.Millisecond)
			mu.Lock()
			fin++
			mu.Unlock()
		}()
	}
	wg.Wait()
	fmt.Printf("  con wg.Wait(): %d de 3\n", fin)

	fmt.Println()
	fmt.Println("== 10.3 · LA TRAMPA 2: dos goroutines tocando el mismo contador ==")
	for intento := 1; intento <= 3; intento++ {
		cuenta := 0
		var g sync.WaitGroup
		for i := 0; i < 1000; i++ {
			g.Add(1)
			go func() {
				defer g.Done()
				for j := 0; j < 100; j++ {
					cuenta++ // ⛔ sin candado
				}
			}()
		}
		g.Wait()
		fmt.Printf("  intento %d: esperaba 100000, salio %d\n", intento, cuenta)
	}
	fmt.Println("  Ni un error. Solo un numero que no cuadra, y distinto cada vez.")
	fmt.Println("  `go run -race` lo detecta. Y un sync.Mutex lo arregla (10.2).")
	fmt.Println()
	fmt.Println("  Nota: el fallo clasico de la variable del for capturada en la")
	fmt.Println("  goroutine YA NO existe desde Go 1.22. Si ves `v := v` en codigo")
	fmt.Println("  viejo, es por eso.")
}
