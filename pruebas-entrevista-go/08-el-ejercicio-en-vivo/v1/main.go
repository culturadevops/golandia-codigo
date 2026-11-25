package main

import (
	"fmt"
	"io"
	"net/http"
	"time"

	"pruebas/banco"
)

// PASO 1. La version secuencial. Funciona, es honesta, y empezar por aqui PUNTUA.
func descargar(urls []string) []string {
	res := make([]string, len(urls))
	for i, u := range urls {
		r, err := http.Get(u)
		if err != nil {
			res[i] = "error: " + err.Error()
			continue
		}
		b, _ := io.ReadAll(r.Body)
		r.Body.Close()
		res[i] = string(b)
	}
	return res
}

func main() {
	s := banco.Nuevo(200 * time.Millisecond)
	defer s.Close()
	urls := banco.URLs(s.URL, 12)

	t := time.Now()
	res := descargar(urls)
	fmt.Printf("12 URLs, 200 ms cada una -> %v\n", time.Since(t).Round(10*time.Millisecond))
	fmt.Println("pico de peticiones a la vez en el servidor:", s.Pico.Load())
	fmt.Println("primera:", res[0], "| ultima:", res[len(res)-1])
}
