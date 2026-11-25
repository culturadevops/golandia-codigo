package main

import (
	"fmt"
	"io"
	"net/http"
	"sync"
	"time"

	"pruebas/banco"
)

// PASO 2. Una goroutine por URL. Rapidisimo... y sin freno.
func descargar(urls []string) []string {
	res := make([]string, len(urls))
	var wg sync.WaitGroup
	for i, u := range urls {
		wg.Add(1)
		go func(i int, u string) { // en Go 1.22+ no haria falta pasar i y u; se pasa por claridad
			defer wg.Done()
			r, err := http.Get(u)
			if err != nil {
				res[i] = "error: " + err.Error()
				return
			}
			b, _ := io.ReadAll(r.Body)
			r.Body.Close()
			res[i] = string(b) // cada goroutine escribe SU indice: no hay carrera
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
	res := descargar(urls)
	fmt.Printf("12 URLs -> %v\n", time.Since(t).Round(10*time.Millisecond))
	fmt.Println("pico de peticiones a la vez:", s.Pico.Load(), " <-- las 12 de golpe. Con 10.000 URLs, esto tumba algo.")
	fmt.Println("resultados:", len(res))
}
