//go:build trampa

//	go test -tags trampa -v ./04_httptest
package httpprueba

import (
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"
)

// La misma petición, pero contra el servidor ENTERO: httptest.NewServer(Rutas()).
func TestVerPedido_Servidor(t *testing.T) {
	srv := httptest.NewServer(Rutas())
	defer srv.Close()
	resp, err := srv.Client().Get(srv.URL + "/pedidos/42")
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	t.Logf("servidor entero: GET /pedidos/42 -> %d", resp.StatusCode)
	if resp.StatusCode != http.StatusOK {
		t.Errorf("código %d; quiero 200", resp.StatusCode)
	}
}

// Un servidor que tarda 2 s. http.DefaultClient no tiene timeout: espera lo que haga falta.
func TestClienteSinTimeout(t *testing.T) {
	lento := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		select {
		case <-time.After(2 * time.Second):
		case <-r.Context().Done():
		}
	}))
	defer lento.Close()
	for _, c := range []struct {
		nombre string
		cli    *http.Client
	}{
		{"http.DefaultClient", http.DefaultClient},
		{"Client{Timeout: 300ms}", &http.Client{Timeout: 300 * time.Millisecond}},
	} {
		t0 := time.Now()
		resp, err := c.cli.Get(lento.URL)
		if err == nil {
			resp.Body.Close()
		}
		t.Logf("%-24s %.2f s  err=%v", c.nombre, time.Since(t0).Seconds(), err != nil)
	}
}

// Cuántas conexiones TCP abre el cliente para 20 peticiones, cerrando o no el Body.
// httptest.Server deja mirar el estado de cada conexión con Config.ConnState.
func TestBodySinCerrar(t *testing.T) {
	for _, cerrar := range []bool{false, true} {
		var nuevas atomic.Int32
		srv := httptest.NewUnstartedServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Write([]byte(`{"id":"cob-1"}`))
		}))
		srv.Config.ConnState = func(_ net.Conn, s http.ConnState) {
			if s == http.StateNew {
				nuevas.Add(1)
			}
		}
		srv.Start()
		cli := &http.Client{Timeout: 5 * time.Second}
		for range 20 {
			resp, err := cli.Get(srv.URL)
			if err != nil {
				t.Fatal(err)
			}
			if cerrar {
				io.Copy(io.Discard, resp.Body)
				resp.Body.Close()
			}
		}
		srv.Close()
		t.Logf("20 peticiones, cerrar Body=%-5v -> conexiones TCP abiertas: %d", cerrar, nuevas.Load())
	}
}
