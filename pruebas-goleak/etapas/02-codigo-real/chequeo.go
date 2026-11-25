// Laboratorio del vídeo 5: un chequeo de salud de N servicios.
// Modos: secuencial | concurrente | carrera | mutex | timeout
// El servidor de prueba se levanta aquí mismo: cada endpoint tarda 200 ms a propósito,
// que es lo que tarda un servicio real que va a base de datos.
package chequeo

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"sync"
	"time"
)

const (
	servicios = 30
	latencia  = 200 * time.Millisecond
)

// Levanta un servidor que simula N servicios lentos.
func servidor() *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/colgado" {
			time.Sleep(30 * time.Second) // el que nunca responde
		}
		time.Sleep(latencia)
		w.WriteHeader(http.StatusOK)
	}))
}

func mirar(url string) bool {
	resp, err := http.Get(url)
	if err != nil {
		return false
	}
	defer resp.Body.Close()
	return resp.StatusCode == http.StatusOK
}

func secuencial(base string) map[string]bool {
	out := map[string]bool{}
	for i := 0; i < servicios; i++ {
		u := fmt.Sprintf("%s/svc-%d", base, i)
		out[u] = mirar(u)
	}
	return out
}

// ⚠️ MAL A PROPÓSITO: varias goroutines escriben el mismo mapa sin protegerlo.
// Es el fallo que enseña el vídeo, y `go run -race` lo detecta.
func carrera(base string) map[string]bool {
	out := map[string]bool{}
	var wg sync.WaitGroup
	for i := 0; i < servicios; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			u := fmt.Sprintf("%s/svc-%d", base, i)
			out[u] = mirar(u) // ← aquí está la carrera
		}(i)
	}
	wg.Wait()
	return out
}

// La versión correcta: el mismo trabajo, con el mapa protegido.
func concurrente(base string) map[string]bool {
	out := map[string]bool{}
	var mu sync.Mutex
	var wg sync.WaitGroup
	for i := 0; i < servicios; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			u := fmt.Sprintf("%s/svc-%d", base, i)
			ok := mirar(u)
			mu.Lock()
			out[u] = ok
			mu.Unlock()
		}(i)
	}
	wg.Wait()
	return out
}

// Con límite de tiempo: el servicio colgado no puede parar el chequeo entero.
func conTimeout(base string, limite time.Duration) (int, int) {
	ctx, cancel := context.WithTimeout(context.Background(), limite)
	defer cancel()
	var mu sync.Mutex
	var wg sync.WaitGroup
	vivos, caidos := 0, 0
	urls := []string{base + "/colgado"}
	for i := 0; i < servicios; i++ {
		urls = append(urls, fmt.Sprintf("%s/svc-%d", base, i))
	}
	for _, u := range urls {
		wg.Add(1)
		go func(u string) {
			defer wg.Done()
			req, _ := http.NewRequestWithContext(ctx, "GET", u, nil)
			resp, err := http.DefaultClient.Do(req)
			mu.Lock()
			defer mu.Unlock()
			if err != nil || resp.StatusCode != http.StatusOK {
				caidos++
				return
			}
			resp.Body.Close()
			vivos++
		}(u)
	}
	wg.Wait()
	return vivos, caidos
}

func noUsar() {
	modo := os.Args[1]
	s := servidor()
	defer s.Close()

	t0 := time.Now()
	switch modo {
	case "secuencial":
		r := secuencial(s.URL)
		fmt.Printf("secuencial: %d servicios en %v\n", len(r), time.Since(t0).Round(time.Millisecond))
	case "concurrente":
		r := concurrente(s.URL)
		fmt.Printf("concurrente: %d servicios en %v\n", len(r), time.Since(t0).Round(time.Millisecond))
	case "carrera":
		r := carrera(s.URL)
		fmt.Printf("carrera: %d servicios en %v\n", len(r), time.Since(t0).Round(time.Millisecond))
	case "timeout":
		v, c := conTimeout(s.URL, 2*time.Second)
		fmt.Printf("timeout: %d vivos, %d caidos, en %v\n", v, c, time.Since(t0).Round(time.Millisecond))
	}
}
