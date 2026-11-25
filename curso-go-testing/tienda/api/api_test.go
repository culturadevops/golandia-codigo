package api

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"curso-testing/tienda"
	"curso-testing/tienda/memoria"
)

// servidor arranca la API ENTERA (con su mux) en un puerto local: así la prueba pasa por las
// rutas de verdad, no solo por el handler (lección 4).
func servidor(t *testing.T) *httptest.Server {
	t.Helper()
	s := &tienda.Servicio{
		Repo:   memoria.Nuevo(),
		Cobros: &memoria.Cobrador{Rechazar: map[string]bool{"moroso": true}},
		Ahora:  func() time.Time { return time.Date(2026, 10, 28, 12, 0, 0, 0, time.UTC) },
	}
	srv := httptest.NewServer(Rutas(s))
	t.Cleanup(srv.Close)
	return srv
}

func TestAPI(t *testing.T) {
	t.Parallel()
	casos := []struct {
		nombre, metodo, ruta, cuerpo string
		codigo                       int
	}{
		{"crear", "POST", "/pedidos", `{"cliente":"ana","items":[{"sku":"TAZA","cantidad":2,"precio_centavos":15000}]}`, 201},
		{"JSON roto", "POST", "/pedidos", `{"cliente":`, 400},
		{"sin cliente", "POST", "/pedidos", `{"items":[]}`, 400},
		{"sin items", "POST", "/pedidos", `{"cliente":"ana","items":[]}`, 422},
		{"cobro rechazado", "POST", "/pedidos", `{"cliente":"moroso","items":[{"sku":"TAZA","cantidad":1,"precio_centavos":100}]}`, 502},
		{"no existe", "GET", "/pedidos/99", "", 404},
		{"id raro", "GET", "/pedidos/abc", "", 400},
		{"método equivocado", "DELETE", "/pedidos/1", "", 405},
	}
	for _, tc := range casos {
		t.Run(tc.nombre, func(t *testing.T) {
			t.Parallel()
			srv := servidor(t)
			req, _ := http.NewRequestWithContext(t.Context(), tc.metodo, srv.URL+tc.ruta,
				strings.NewReader(tc.cuerpo))
			resp, err := srv.Client().Do(req)
			if err != nil {
				t.Fatal(err)
			}
			defer resp.Body.Close()
			io.Copy(io.Discard, resp.Body)
			if resp.StatusCode != tc.codigo {
				t.Errorf("%s %s = %d; quiero %d", tc.metodo, tc.ruta, resp.StatusCode, tc.codigo)
			}
		})
	}
}

func TestCrearYVer(t *testing.T) {
	t.Parallel()
	srv := servidor(t)
	resp, err := srv.Client().Post(srv.URL+"/pedidos", "application/json", strings.NewReader(
		`{"cliente":"ana","items":[{"sku":"TAZA","cantidad":2,"precio_centavos":15000}]}`))
	if err != nil {
		t.Fatal(err)
	}
	var creado tienda.Pedido
	json.NewDecoder(resp.Body).Decode(&creado)
	resp.Body.Close()

	resp, err = srv.Client().Get(srv.URL + "/pedidos/1")
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	var visto tienda.Pedido
	json.NewDecoder(resp.Body).Decode(&visto)
	// 300,00 - 10 % de primer pedido + 49,00 de envío = 319,00
	if visto.Total != 31900 || visto.Estado != "pagado" || visto.ID != creado.ID {
		t.Errorf("visto = %+v", visto)
	}
}
