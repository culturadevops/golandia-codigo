package banco

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
