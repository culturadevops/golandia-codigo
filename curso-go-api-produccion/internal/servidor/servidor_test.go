package servidor_test

import (
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"tareas/internal/bd"
	"tareas/internal/cuentas"
	"tareas/internal/pruebabd"
	"tareas/internal/servidor"
	"tareas/internal/tareas"
	"tareas/migraciones"
)

// TestDePuntaAPunta usa el handler completo (middlewares incluidos) contra Postgres.
func TestDePuntaAPunta(t *testing.T) {
	pool := pruebabd.Abrir(t)
	salud := &servidor.Salud{BD: bd.Sonda{Pool: pool}, Esperada: migraciones.Ultima()}
	srv := httptest.NewServer(servidor.Rutas(servidor.Deps{
		Log:     slog.New(slog.NewJSONHandler(io.Discard, nil)),
		Cuentas: cuentas.Nuevo(pool, 10),
		Tareas:  tareas.Postgres{BD: pool},
		Salud:   salud,
		Plazo:   5 * time.Second,
	}))
	defer srv.Close()

	pedir := func(metodo, ruta, cuerpo, token string, quiero int) map[string]any {
		t.Helper()
		req, _ := http.NewRequest(metodo, srv.URL+ruta, strings.NewReader(cuerpo))
		if token != "" {
			req.Header.Set("Authorization", "Bearer "+token)
		}
		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			t.Fatal(err)
		}
		defer resp.Body.Close()
		var m map[string]any
		_ = json.NewDecoder(resp.Body).Decode(&m)
		if resp.StatusCode != quiero {
			t.Fatalf("%s %s: %d, quería %d: %v", metodo, ruta, resp.StatusCode, quiero, m)
		}
		return m
	}

	pedir("GET", "/salud/listo", "", "", 200)
	pedir("POST", "/usuarios", `{"email":"ana@ejemplo.com","clave":"una clave larga"}`, "", 201)
	tok := pedir("POST", "/sesiones", `{"email":"ana@ejemplo.com","clave":"una clave larga"}`, "", 201)["token"].(string)

	pedir("GET", "/tareas", "", "", 401)
	pedir("POST", "/tareas", `{"titulo":"probar la API"}`, tok, 201)
	l := pedir("GET", "/tareas", "", tok, 200)
	if n := len(l["tareas"].([]any)); n != 1 {
		t.Fatalf("esperaba 1 tarea, hay %d", n)
	}
	pedir("PATCH", "/tareas/1", `{"hecha":true}`, tok, 200)
	pedir("DELETE", "/tareas/1", "", tok, 204)
	pedir("GET", "/tareas/1", "", tok, 404)
	pedir("DELETE", "/sesiones", "", tok, 204)
	pedir("GET", "/tareas", "", tok, 401)

	salud.Apagando()
	pedir("GET", "/salud/listo", "", "", 503)
	pedir("GET", "/salud/vivo", "", "", 200)
}
