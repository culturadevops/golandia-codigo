package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"sync"
	"time"

	"pruebas/banco"
)

type Resultado struct {
	URL   string
	Datos string
	Err   error
}

// PASO 4. Cancelable y con los errores recogidos, ninguno perdido.
func descargar(ctx context.Context, urls []string, k int) []Resultado {
	res := make([]Resultado, len(urls))
	sem := make(chan struct{}, k)
	var wg sync.WaitGroup
	for i, u := range urls {
		wg.Add(1)
		go func(i int, u string) {
			defer wg.Done()
			res[i].URL = u
			select {
			case sem <- struct{}{}:
				defer func() { <-sem }()
			case <-ctx.Done(): // si nos cancelan mientras esperamos turno, ni empezamos
				res[i].Err = ctx.Err()
				return
			}
			req, err := http.NewRequestWithContext(ctx, "GET", u, nil)
			if err != nil {
				res[i].Err = err
				return
			}
			r, err := http.DefaultClient.Do(req)
			if err != nil {
				res[i].Err = err
				return
			}
			defer r.Body.Close()
			b, err := io.ReadAll(r.Body)
			res[i].Datos, res[i].Err = string(b), err
		}(i, u)
	}
	wg.Wait()
	return res
}

func resumen(titulo string, res []Resultado, tardo time.Duration, pico int64) {
	ok, cancelados, otros := 0, 0, 0
	for _, r := range res {
		switch {
		case r.Err == nil:
			ok++
		case errors.Is(r.Err, context.Canceled), errors.Is(r.Err, context.DeadlineExceeded):
			cancelados++
		default:
			otros++
		}
	}
	fmt.Printf("%-22s %6v  ok=%-3d cancelados=%-3d otros=%-3d pico=%d\n",
		titulo, tardo.Round(10*time.Millisecond), ok, cancelados, otros, pico)
}

func main() {
	// a) sin cancelar
	s := banco.Nuevo(200 * time.Millisecond)
	urls := banco.URLs(s.URL, 12)
	t := time.Now()
	resumen("completo:", descargar(context.Background(), urls, 4), time.Since(t), s.Pico.Load())
	s.Close()

	// b) con un plazo que corta a la mitad
	s2 := banco.Nuevo(200 * time.Millisecond)
	urls2 := banco.URLs(s2.URL, 12)
	ctx, cancel := context.WithTimeout(context.Background(), 350*time.Millisecond)
	defer cancel()
	t = time.Now()
	resumen("cancelado a 350ms:", descargar(ctx, urls2, 4), time.Since(t), s2.Pico.Load())
	s2.Close()

	fmt.Println()
	fmt.Println("Ni un error perdido y nada colgado. Ese es el estado final que buscaba el entrevistador.")
}
