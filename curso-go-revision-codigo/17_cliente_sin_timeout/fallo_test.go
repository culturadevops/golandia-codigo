package pedidos

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestCotizarConUnaPaqueteriaColgada(t *testing.T) {
	colgada := make(chan struct{})
	srv := httptest.NewServer(http.HandlerFunc(func(_ http.ResponseWriter, _ *http.Request) {
		<-colgada // acepta la conexión y nunca contesta
	}))
	t.Cleanup(srv.Close)
	t.Cleanup(func() { close(colgada) })

	listo := make(chan error, 1)
	inicio := time.Now()
	go func() { _, err := Cotizar(srv.URL); listo <- err }()
	select {
	case err := <-listo:
		t.Logf("paquetería colgada -> Cotizar volvió a los %v: %v", time.Since(inicio).Round(100*time.Millisecond), err)
	case <-time.After(3 * time.Second):
		t.Fatalf("paquetería colgada -> Cotizar sigue esperando tras 3s (y sin Timeout, esperaría para siempre)")
	}
}
