# -*- coding: utf-8 -*-
import sys
from pathlib import Path
sys.path.insert(0, str(Path(__file__).resolve().parent))
from probar import episodio

# Un servidor de mentira compartido por las cuatro versiones: 200 ms por peticion y
# cuenta cuantas atiende a la vez. Asi el limite se DEMUESTRA, no se promete.
BANCO = '''package banco

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"time"
)

type Servidor struct {
	*httptest.Server
	EnVuelo atomic.Int64
	Pico    atomic.Int64
}

// Nuevo levanta un servidor que tarda `retardo` y registra el pico de concurrencia.
func Nuevo(retardo time.Duration) *Servidor {
	s := &Servidor{}
	s.Server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		n := s.EnVuelo.Add(1)
		for {
			p := s.Pico.Load()
			if n <= p || s.Pico.CompareAndSwap(p, n) {
				break
			}
		}
		defer s.EnVuelo.Add(-1)
		select {
		case <-time.After(retardo):
			fmt.Fprintf(w, "ok %s", r.URL.Path)
		case <-r.Context().Done():
		}
	}))
	return s
}

func URLs(base string, n int) []string {
	out := make([]string, n)
	for i := range out {
		out[i] = fmt.Sprintf("%s/doc-%02d", base, i+1)
	}
	return out
}
'''

V1 = '''package main

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
	fmt.Printf("12 URLs, 200 ms cada una -> %v\\n", time.Since(t).Round(10*time.Millisecond))
	fmt.Println("pico de peticiones a la vez en el servidor:", s.Pico.Load())
	fmt.Println("primera:", res[0], "| ultima:", res[len(res)-1])
}
'''

V2 = '''package main

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
	fmt.Printf("12 URLs -> %v\\n", time.Since(t).Round(10*time.Millisecond))
	fmt.Println("pico de peticiones a la vez:", s.Pico.Load(), " <-- las 12 de golpe. Con 10.000 URLs, esto tumba algo.")
	fmt.Println("resultados:", len(res))
}
'''

V3 = '''package main

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
	fmt.Printf("12 URLs con maximo 4 a la vez -> %v\\n", time.Since(t).Round(10*time.Millisecond))
	fmt.Println("pico de peticiones a la vez:", s.Pico.Load(), " <-- respetado")
	fmt.Println()
	fmt.Println("3 tandas de 4 x 200 ms = 600 ms. El numero cuadra, y eso es lo que hay que decir en voz alta.")
}
'''

V4 = '''package main

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
	fmt.Printf("%-22s %6v  ok=%-3d cancelados=%-3d otros=%-3d pico=%d\\n",
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
'''

TEST = '''package descarga

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
'''

episodio(
    "08-el-ejercicio-en-vivo",
    "Episodio 8/8 — el ejercicio en vivo: N URLs, maximo K a la vez, cancelable",
    [
        {"nombre": "Paso 1. La version secuencial (empezar por aqui PUNTUA)",
         "pregunta": "Descarga N URLs. Sin mas. Que escribes primero?",
         "archivos": {"banco/banco.go": BANCO, "v1/main.go": V1},
         "correr": [(["go", "run", "./v1"], False)]},
        {"nombre": "Paso 2. Una goroutine por URL: rapido y sin freno",
         "pregunta": "Ahora hazlo concurrente.",
         "archivos": {"v2/main.go": V2},
         "correr": [(["go", "run", "./v2"], False)],
         "nota": "Aqui el entrevistador pregunta: y si son 10.000 URLs? Ese es el paso 3."},
        {"nombre": "Paso 3. El semaforo: maximo K a la vez",
         "pregunta": "Limitalo a K simultaneas.",
         "archivos": {"v3/main.go": V3},
         "correr": [(["go", "run", "./v3"], False)]},
        {"nombre": "Paso 4. Cancelable y sin perder un solo error",
         "pregunta": "Y si el usuario cierra la pestana a mitad?",
         "archivos": {"v4/main.go": V4},
         "correr": [(["go", "run", "-race", "./v4"], False)]},
        {"nombre": "Paso 5. El test que lo DEMUESTRA (el que casi nadie escribe)",
         "pregunta": "Como sabes que el limite se respeta de verdad?",
         "archivos": {"descarga/descarga_test.go": TEST},
         "correr": [(["go", "test", "-v", "-race", "./descarga"], False)],
         "nota": "Prometer el limite es facil. Contarlo desde el otro lado con un atomico y afirmarlo\nen un test es lo que separa al que pasa la ronda del que la pasa bien."},
    ])
