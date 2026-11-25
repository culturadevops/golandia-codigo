package descarga

import (
	"context"
	"io"
	"net/http"
	"sync"
	"testing"
	"time"

	"pruebas/banco"
)

func descargar(ctx context.Context, urls []string, k int) []string {
	res := make([]string, len(urls))
	sem := make(chan struct{}, k)
	var wg sync.WaitGroup
	for i, u := range urls {
		wg.Add(1)
		go func(i int, u string) {
			defer wg.Done()
			select {
			case sem <- struct{}{}:
				defer func() { <-sem }()
			case <-ctx.Done():
				return
			}
			req, _ := http.NewRequestWithContext(ctx, "GET", u, nil)
			r, err := http.DefaultClient.Do(req)
			if err != nil {
				return
			}
			defer r.Body.Close()
			b, _ := io.ReadAll(r.Body)
			res[i] = string(b)
		}(i, u)
	}
	wg.Wait()
	return res
}

// EL PASO QUE CONTRATA: no prometer el limite, DEMOSTRARLO.
func TestRespetaElLimite(t *testing.T) {
	for _, k := range []int{1, 3, 5} {
		s := banco.Nuevo(30 * time.Millisecond)
		urls := banco.URLs(s.URL, 20)
		descargar(context.Background(), urls, k)
		if pico := s.Pico.Load(); pico > int64(k) {
			t.Errorf("k=%d: el servidor llego a ver %d peticiones a la vez", k, pico)
		} else {
			t.Logf("k=%d -> pico observado %d  OK", k, pico)
		}
		s.Close()
	}
}

func TestDevuelveTodo(t *testing.T) {
	s := banco.Nuevo(10 * time.Millisecond)
	defer s.Close()
	urls := banco.URLs(s.URL, 20)
	res := descargar(context.Background(), urls, 4)
	if len(res) != len(urls) {
		t.Fatalf("esperaba %d resultados, hay %d", len(urls), len(res))
	}
	for i, r := range res {
		if r == "" {
			t.Errorf("el resultado %d vino vacio", i)
		}
	}
}

func TestSeCancela(t *testing.T) {
	s := banco.Nuevo(200 * time.Millisecond)
	defer s.Close()
	urls := banco.URLs(s.URL, 20)
	ctx, cancel := context.WithTimeout(context.Background(), 250*time.Millisecond)
	defer cancel()
	inicio := time.Now()
	descargar(ctx, urls, 2)
	if d := time.Since(inicio); d > time.Second {
		t.Errorf("no se corto: tardo %v", d)
	} else {
		t.Logf("cancelado en %v en vez de los ~2s que tardaria entero", d.Round(10*time.Millisecond))
	}
}
