package pedidos

import (
	"net"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
)

func TestDisponibleReutilizaLaConexion(t *testing.T) {
	var nuevas atomic.Int32
	srv := httptest.NewUnstartedServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`{"sku":"taza","stock":12}`))
	}))
	srv.Config.ConnState = func(_ net.Conn, s http.ConnState) {
		if s == http.StateNew {
			nuevas.Add(1)
		}
	}
	srv.Start()
	defer srv.Close()

	for range 50 {
		if _, err := Disponible(srv.Client(), srv.URL); err != nil {
			t.Fatal(err)
		}
	}
	if n := nuevas.Load(); n > 2 {
		t.Fatalf("50 consultas al almacén -> %d conexiones TCP nuevas", n)
	}
	t.Logf("50 consultas al almacén -> %d conexión TCP", nuevas.Load())
}
