package main

import (
	"fmt"
	"io"
	"net/http"
	"sync"
	"time"

	"pruebas/banco"
)

// PASO 3. El limite: un canal con buffer haciendo de semaforo.
func descargar(urls []string, k int) []string {
	res := make([]string, len(urls))
	sem := make(chan struct{}, k)
	var wg sync.WaitGroup
	for i, u := range urls {
		wg.Add(1)
		go func(i int, u string) {
			defer wg.Done()
			sem <- struct{}{}        // coge sitio (bloquea si ya hay k dentro)
			defer func() { <-sem }() // suelta
			r, err := http.Get(u)
			if err != nil {
				res[i] = "error: " + err.Error()
				return
			}
			b, _ := io.ReadAll(r.Body)
			r.Body.Close()
			res[i] = string(b)
		}(i, u)
	}
	wg.Wait()
	return res
}

func main() {
	s := banco.Nuevo(200 * time.Millisecond)
	defer s.Close()
	urls := banco.URLs(s.URL, 12)

	t := time.Now()
	descargar(urls, 4)
	fmt.Printf("12 URLs con maximo 4 a la vez -> %v\n", time.Since(t).Round(10*time.Millisecond))
	fmt.Println("pico de peticiones a la vez:", s.Pico.Load(), " <-- respetado")
	fmt.Println()
	fmt.Println("3 tandas de 4 x 200 ms = 600 ms. El numero cuadra, y eso es lo que hay que decir en voz alta.")
}
