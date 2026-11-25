// Package comun tiene lo que comparten todas las lecciones: revisar un endpoint HTTP.
//
// Es la parte "de negocio" de vigia, y a propósito es pequeña: el curso va de la CLI
// que la envuelve (flags, configuración, salida, códigos, señales, tests y release).
package comun

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"sync"
	"time"
)

// Resultado es lo que sabemos de un endpoint después de pedirlo una vez.
type Resultado struct {
	URL      string
	Estado   int           // código HTTP; 0 si ni siquiera hubo respuesta
	Latencia time.Duration // desde que salió la petición hasta que llegó la respuesta (o el error)
	Err      error
}

// OK dice si el endpoint está arriba: respondió, y con un 2xx o 3xx.
func (r Resultado) OK() bool {
	return r.Err == nil && r.Estado >= 200 && r.Estado < 400
}

// Motivo es el error en pocas palabras, para una línea de terminal.
func (r Resultado) Motivo() string {
	switch {
	case r.Err == nil && r.OK():
		return ""
	case r.Err == nil:
		return http.StatusText(r.Estado)
	case errors.Is(r.Err, context.DeadlineExceeded):
		return "timeout"
	case errors.Is(r.Err, context.Canceled):
		return "cancelado"
	default:
		return r.Err.Error()
	}
}

// Revisar pide la URL una vez, con un tiempo máximo, y respetando ctx: si ctx se
// cancela (Ctrl+C), la petición se corta en el acto.
func Revisar(ctx context.Context, url string, timeout time.Duration) Resultado {
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return Resultado{URL: url, Err: err}
	}
	ini := time.Now()
	resp, err := http.DefaultClient.Do(req)
	lat := time.Since(ini)
	if err != nil {
		return Resultado{URL: url, Latencia: lat, Err: err}
	}
	_, _ = io.Copy(io.Discard, resp.Body)
	resp.Body.Close()
	return Resultado{URL: url, Estado: resp.StatusCode, Latencia: lat}
}

// RevisarTodas revisa las URLs con, como mucho, `concurrencia` peticiones a la vez.
// Los resultados salen en el mismo orden que las URLs.
func RevisarTodas(ctx context.Context, urls []string, timeout time.Duration, concurrencia int) []Resultado {
	if concurrencia < 1 {
		concurrencia = 1
	}
	res := make([]Resultado, len(urls))
	turno := make(chan struct{}, concurrencia)
	var wg sync.WaitGroup
	for i, u := range urls {
		wg.Add(1)
		go func() {
			defer wg.Done()
			turno <- struct{}{}
			defer func() { <-turno }()
			res[i] = Revisar(ctx, u, timeout)
		}()
	}
	wg.Wait()
	return res
}

// Linea es el resultado para una persona: estado, código, latencia y URL.
func Linea(r Resultado) string {
	estado := "ARRIBA"
	if !r.OK() {
		estado = "CAIDO "
	}
	codigo := "---"
	if r.Estado != 0 {
		codigo = fmt.Sprint(r.Estado)
	}
	l := fmt.Sprintf("%s  %s  %7.1f ms  %s", estado, codigo, MS(r.Latencia), r.URL)
	if m := r.Motivo(); m != "" {
		l += "  (" + m + ")"
	}
	return l
}

// MS es la latencia en milisegundos, con decimales.
func MS(d time.Duration) float64 {
	return float64(d.Microseconds()) / 1000
}
