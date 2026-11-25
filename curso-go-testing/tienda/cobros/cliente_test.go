package cobros

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"
)

// pasarela levanta un servidor HTTP de verdad, en un puerto local, que contesta con los
// códigos de la lista en orden. httptest.NewServer + t.Cleanup(srv.Close).
func pasarela(t *testing.T, codigos ...int) (*httptest.Server, *atomic.Int32) {
	t.Helper()
	var n atomic.Int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		i := int(n.Add(1)) - 1
		c := codigos[min(i, len(codigos)-1)]
		w.WriteHeader(c)
		if c == http.StatusCreated {
			w.Write([]byte(`{"id":"cob-1"}`))
		} else {
			w.Write([]byte(`{"error":"x"}`))
		}
	}))
	t.Cleanup(srv.Close)
	return srv, &n
}

func cliente(srv *httptest.Server) *Cliente {
	return &Cliente{Base: srv.URL, HTTP: srv.Client(), Espera: time.Millisecond}
}

func TestCobrar(t *testing.T) {
	t.Parallel()
	casos := []struct {
		nombre   string
		codigos  []int
		llamadas int32
		quiereOK bool
	}{
		{"a la primera", []int{201}, 1, true},
		{"dos 503 y luego bien", []int{503, 503, 201}, 3, true},
		{"siempre 503: se rinde a los 4", []int{503}, 4, false},
		{"un 402 no se reintenta", []int{402}, 1, false},
	}
	for _, tc := range casos {
		t.Run(tc.nombre, func(t *testing.T) {
			t.Parallel()
			srv, n := pasarela(t, tc.codigos...)
			id, err := cliente(srv).Cobrar(t.Context(), "ana", 1250)
			if (err == nil) != tc.quiereOK {
				t.Fatalf("err = %v", err)
			}
			if tc.quiereOK && id != "cob-1" {
				t.Errorf("id = %q", id)
			}
			if got := n.Load(); got != tc.llamadas {
				t.Errorf("llamadas = %d; quiero %d", got, tc.llamadas)
			}
		})
	}
}

func TestCobrarRespetaElContext(t *testing.T) {
	t.Parallel()
	srv, _ := pasarela(t, 503)
	c := cliente(srv)
	c.Espera = time.Hour
	ctx, cancel := context.WithTimeout(t.Context(), 50*time.Millisecond)
	defer cancel()
	_, err := c.Cobrar(ctx, "ana", 1250)
	if !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("err = %v; quiero DeadlineExceeded", err)
	}
}
