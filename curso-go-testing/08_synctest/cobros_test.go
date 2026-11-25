package relojes

import (
	"context"
	"errors"
	"io"
	"net/http"
	"strings"
	"sync/atomic"
	"testing"
	"testing/synctest"
	"time"

	"curso-testing/tienda/cobros"
)

// pasarela es un http.RoundTripper en memoria: sin red, para que todo quede DENTRO de la
// burbuja de synctest (una goroutine esperando a un socket de verdad no cuenta como dormida).
type pasarela struct {
	codigos []int
	n       atomic.Int32
}

func (p *pasarela) RoundTrip(*http.Request) (*http.Response, error) {
	i := int(p.n.Add(1)) - 1
	c := p.codigos[min(i, len(p.codigos)-1)]
	cuerpo := `{"error":"ocupado"}`
	if c == http.StatusCreated {
		cuerpo = `{"id":"cob-7"}`
	}
	return &http.Response{StatusCode: c, Body: io.NopCloser(strings.NewReader(cuerpo)),
		Header: http.Header{}}, nil
}

func cliente(p *pasarela) *cobros.Cliente {
	return &cobros.Cliente{Base: "http://pasarela", HTTP: &http.Client{Transport: p, Timeout: 10 * time.Second}}
}

// Dentro de synctest.Test el reloj es falso: avanza solo cuando todas las goroutines de la
// burbuja están dormidas. Tres 503 y un 201 son 1 + 2 + 4 = 7 s... exactos, y sin esperarlos.
func TestReintentos(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		p := &pasarela{codigos: []int{503, 503, 503, 201}}
		t0 := time.Now()
		id, err := cliente(p).Cobrar(t.Context(), "ana", 1250)
		if err != nil || id != "cob-7" {
			t.Fatalf("id %q, err %v", id, err)
		}
		dentro := time.Since(t0)
		t.Logf("llamadas: %d · tiempo dentro de la burbuja: %v", p.n.Load(), dentro)
		if dentro != 7*time.Second {
			t.Errorf("quiero exactamente 7s; fueron %v", dentro)
		}
	})
}

// Se rinde al cuarto intento: 1 + 2 + 4 = 7 s y devuelve el último error.
func TestSeRinde(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		p := &pasarela{codigos: []int{503}}
		t0 := time.Now()
		_, err := cliente(p).Cobrar(t.Context(), "ana", 1250)
		var e *cobros.ErrHTTP
		if !errors.As(err, &e) || e.Codigo != 503 || p.n.Load() != 4 {
			t.Fatalf("err %v, llamadas %d", err, p.n.Load())
		}
		t.Logf("se rinde tras %d llamadas, a los %v", p.n.Load(), time.Since(t0))
	})
}

// El context vence a los 2,5 s, en mitad de la segunda espera (de 1 s a 3 s).
// El cliente de verdad usa select: sale a los 2,5 s exactos.
func TestCancelaDuranteLaEspera(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		ctx, cancel := context.WithTimeout(t.Context(), 2500*time.Millisecond)
		defer cancel()
		t0 := time.Now()
		_, err := cliente(&pasarela{codigos: []int{503}}).Cobrar(ctx, "ana", 1250)
		t.Logf("cobros.Cliente: sale a los %v con %v", time.Since(t0), err)
		if !errors.Is(err, context.DeadlineExceeded) || time.Since(t0) != 2500*time.Millisecond {
			t.Errorf("quiero DeadlineExceeded a los 2.5s")
		}
	})
}
