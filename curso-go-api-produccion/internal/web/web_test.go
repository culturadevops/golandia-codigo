package web

import (
	"bytes"
	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

type tarea struct {
	Titulo    string `json:"titulo"`
	Prioridad int    `json:"prioridad"`
}

func leerCon(cuerpo string) error {
	r := httptest.NewRequest("POST", "/", strings.NewReader(cuerpo))
	var v tarea
	return Leer(httptest.NewRecorder(), r, &v)
}

func TestLeer(t *testing.T) {
	casos := []struct{ nombre, cuerpo, codigo string }{
		{"bien", `{"titulo":"a","prioridad":1}`, ""},
		{"campo mal escrito", `{"titulo":"a","priorida":1}`, "peticion_invalida"},
		{"dos objetos", `{"titulo":"a"}{"titulo":"b"}`, "peticion_invalida"},
		{"tipo mal", `{"titulo":5}`, "peticion_invalida"},
		{"vacío", ``, "peticion_invalida"},
		{"enorme", `{"titulo":"` + strings.Repeat("x", MaxCuerpo) + `"}`, "demasiado_grande"},
	}
	for _, c := range casos {
		t.Run(c.nombre, func(t *testing.T) {
			err := leerCon(c.cuerpo)
			got := ""
			if e, ok := err.(*Error); ok {
				got = e.Codigo
			} else if err != nil {
				got = "otro: " + err.Error()
			}
			if got != c.codigo {
				t.Fatalf("código %q, esperaba %q", got, c.codigo)
			}
		})
	}
}

func TestRegistroVeElEstadoReal(t *testing.T) {
	var buf bytes.Buffer
	log := slog.New(slog.NewJSONHandler(&buf, nil))
	h := Cadena(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte("hola")) // sin WriteHeader: Go manda 200
	}), ConID(log), Registro)
	h.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest("GET", "/", nil))
	if !strings.Contains(buf.String(), `"estado":200`) {
		t.Fatalf("el log no dice 200: %s", buf.String())
	}
}

func TestRecuperarDevuelveJSONConID(t *testing.T) {
	log := slog.New(slog.NewJSONHandler(&bytes.Buffer{}, nil))
	h := Cadena(http.HandlerFunc(func(http.ResponseWriter, *http.Request) { panic("boom") }),
		ConID(log), Recuperar)
	rec := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/", nil)
	req.Header.Set("X-Request-Id", "abc-123")
	h.ServeHTTP(rec, req)
	var s sobre
	_ = json.Unmarshal(rec.Body.Bytes(), &s)
	if rec.Code != 500 || s.Error == nil || s.Error.Codigo != "interno" || s.RequestID != "abc-123" {
		t.Fatalf("estado %d, cuerpo %s", rec.Code, rec.Body)
	}
}

func TestErroresDelMuxEnJSON(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /x", func(http.ResponseWriter, *http.Request) {})
	h := ErroresDelMux(mux)
	for ruta, estado := range map[string]int{"GET /nada": 404, "DELETE /x": 405} {
		m, p, _ := strings.Cut(ruta, " ")
		rec := httptest.NewRecorder()
		h.ServeHTTP(rec, httptest.NewRequest(m, p, nil))
		if rec.Code != estado || rec.Header().Get("Content-Type") != "application/json" {
			t.Errorf("%s: %d %s %s", ruta, rec.Code, rec.Header().Get("Content-Type"), rec.Body)
		}
	}
}
