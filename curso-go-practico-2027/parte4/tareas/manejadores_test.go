package main

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// pedir manda una petición a la API (con el almacén en memoria) sin abrir ningún puerto.
func pedir(t *testing.T, api http.Handler, metodo, ruta, cuerpo string) *httptest.ResponseRecorder {
	t.Helper()
	req := httptest.NewRequest(metodo, ruta, strings.NewReader(cuerpo))
	rec := httptest.NewRecorder()
	api.ServeHTTP(rec, req)
	return rec
}

func TestCrearTarea(t *testing.T) {
	casos := []struct {
		nombre string
		cuerpo string
		estado int
	}{
		{"válida", `{"titulo":"comprar pan"}`, http.StatusCreated},
		{"título vacío", `{"titulo":"  "}`, http.StatusUnprocessableEntity},
		{"JSON roto", `{"titulo":`, http.StatusBadRequest},
		{"campo desconocido", `{"titel":"pan"}`, http.StatusBadRequest},
		{"título largo", `{"titulo":"` + strings.Repeat("x", 121) + `"}`, http.StatusUnprocessableEntity},
	}
	for _, c := range casos {
		t.Run(c.nombre, func(t *testing.T) {
			api := (&API{almacen: NuevaMemoria()}).Rutas()
			rec := pedir(t, api, "POST", "/tareas", c.cuerpo)
			if rec.Code != c.estado {
				t.Errorf("estado = %d, quiero %d; cuerpo: %s", rec.Code, c.estado, rec.Body)
			}
		})
	}
}

func TestCrearDevuelveLocation(t *testing.T) {
	api := (&API{almacen: NuevaMemoria()}).Rutas()
	rec := pedir(t, api, "POST", "/tareas", `{"titulo":"comprar pan"}`)
	if got := rec.Header().Get("Location"); got != "/tareas/1" {
		t.Errorf("Location = %q, quiero /tareas/1", got)
	}
}

func TestRutas(t *testing.T) {
	api := (&API{almacen: NuevaMemoria()}).Rutas()
	casos := []struct {
		metodo, ruta string
		estado       int
	}{
		{"GET", "/tareas/99", http.StatusNotFound},
		{"GET", "/tareas/abc", http.StatusBadRequest},
		{"PUT", "/tareas", http.StatusMethodNotAllowed},
		{"DELETE", "/tareas/1", http.StatusNotFound},
	}
	for _, c := range casos {
		t.Run(c.metodo+" "+c.ruta, func(t *testing.T) {
			if rec := pedir(t, api, c.metodo, c.ruta, ""); rec.Code != c.estado {
				t.Errorf("estado = %d, quiero %d", rec.Code, c.estado)
			}
		})
	}
}

// Con httptest.NewServer hay un servidor de verdad, en un puerto libre, y un cliente HTTP real.
func TestFlujoCompleto(t *testing.T) {
	srv := httptest.NewServer((&API{almacen: NuevaMemoria()}).Rutas())
	defer srv.Close()
	flujo(t, srv.URL)
}

// flujo crea, marca, lee y borra una tarea. Lo usan la prueba en memoria y la de Postgres.
func flujo(t *testing.T, base string) {
	t.Helper()
	pasos := []struct {
		metodo, ruta, cuerpo string
		estado               int
	}{
		{"POST", "/tareas", `{"titulo":"comprar pan"}`, 201},
		{"PATCH", "/tareas/1", `{"hecha":true}`, 200},
		{"GET", "/tareas/1", "", 200},
		{"DELETE", "/tareas/1", "", 204},
		{"GET", "/tareas/1", "", 404},
	}
	for _, p := range pasos {
		req, _ := http.NewRequest(p.metodo, base+p.ruta, strings.NewReader(p.cuerpo))
		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			t.Fatal(err)
		}
		resp.Body.Close()
		if resp.StatusCode != p.estado {
			t.Fatalf("%s %s: estado = %d, quiero %d", p.metodo, p.ruta, resp.StatusCode, p.estado)
		}
	}
}
