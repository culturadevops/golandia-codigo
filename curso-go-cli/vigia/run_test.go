package main

import (
	"bytes"
	"context"
	"flag"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"
)

var actualizar = flag.Bool("update", false, "reescribe los archivos .golden con la salida actual")

// banco de pruebas en el propio proceso: /ok, /roto y /lento.
func banco(t *testing.T) string {
	t.Helper()
	mux := http.NewServeMux()
	mux.HandleFunc("/ok", func(w http.ResponseWriter, r *http.Request) {})
	mux.HandleFunc("/roto", func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(500) })
	mux.HandleFunc("/lento", func(w http.ResponseWriter, r *http.Request) {
		select {
		case <-time.After(2 * time.Second):
		case <-r.Context().Done():
		}
	})
	srv := httptest.NewServer(mux)
	t.Cleanup(srv.Close)
	return srv.URL
}

// correr llama a run como lo haría main, pero con un entorno inventado: nada del
// os.Getenv de quien corre los tests se cuela aquí.
func correr(args []string, stdin string, env map[string]string) (int, string, string) {
	var out, errb bytes.Buffer
	e := Entorno{
		Stdin: strings.NewReader(stdin), Stdout: &out, Stderr: &errb,
		Getenv: func(k string) string { return env[k] },
	}
	codigo := run(context.Background(), args, e)
	return codigo, out.String(), errb.String()
}

func TestCodigosDeSalida(t *testing.T) {
	srv := banco(t)
	casos := []struct {
		nombre   string
		args     []string
		stdin    string
		env      map[string]string
		codigo   int
		enStdout string
		enStderr string
	}{
		{"todo arriba", []string{"revisar", srv + "/ok"}, "", nil, 0, "ARRIBA", "revisando 1"},
		{"uno caído", []string{"revisar", srv + "/ok", srv + "/roto"}, "", nil, 1, "CAIDO", "1 de 2 caídos"},
		{"timeout", []string{"revisar", "-timeout", "100ms", srv + "/lento"}, "", nil, 1, "(timeout)", ""},
		{"sin URLs", []string{"revisar"}, "", nil, 2, "", "no hay URLs"},
		{"flag mal", []string{"revisar", "-timeout", "3", srv + "/ok"}, "", nil, 2, "", "invalid value"},
		{"entorno mal", []string{"revisar", srv + "/ok"}, "", map[string]string{"VIGIA_TIMEOUT": "3"}, 2, "", "VIGIA_TIMEOUT"},
		{"urls por stdin", []string{"revisar", "-"}, "# comentario\n" + srv + "/ok\n\n", nil, 0, "ARRIBA", "revisando 1"},
		{"subcomando mal", []string{"revisr"}, "", nil, 2, "", "desconocido"},
		{"ayuda", []string{"revisar", "-h"}, "", nil, 0, "", "-solo-caidos"},
	}
	for _, c := range casos {
		t.Run(c.nombre, func(t *testing.T) {
			codigo, out, errb := correr(c.args, c.stdin, c.env)
			if codigo != c.codigo {
				t.Errorf("código %d, quería %d\nstdout: %s\nstderr: %s", codigo, c.codigo, out, errb)
			}
			if !strings.Contains(out, c.enStdout) {
				t.Errorf("stdout sin %q:\n%s", c.enStdout, out)
			}
			if !strings.Contains(errb, c.enStderr) {
				t.Errorf("stderr sin %q:\n%s", c.enStderr, errb)
			}
		})
	}
}

// En -json, stdout tiene que ser JSON y nada más: el progreso va a stderr.
func TestJSONLimpio(t *testing.T) {
	srv := banco(t)
	_, out, _ := correr([]string{"revisar", "-json", "-solo-caidos", srv + "/ok"}, "", nil)
	if strings.TrimSpace(out) != "[]" {
		t.Fatalf("sin caídos quería [], salió %q", out)
	}
}

func TestPrecedencia(t *testing.T) {
	dir := t.TempDir()
	t.Chdir(dir)
	os.WriteFile("vigia.json", []byte(`{"timeout":"5s"}`), 0o644)
	casos := []struct {
		args []string
		env  map[string]string
		want string
	}{
		{nil, nil, "5s     (archivo vigia.json)"},
		{nil, map[string]string{"VIGIA_TIMEOUT": "3s"}, "3s     (entorno VIGIA_TIMEOUT)"},
		{[]string{"-timeout", "10s"}, map[string]string{"VIGIA_TIMEOUT": "3s"}, "10s    (flag -timeout)"},
	}
	for _, c := range casos {
		_, out, errb := correr(append([]string{"config"}, c.args...), "", c.env)
		if !strings.Contains(out, c.want) {
			t.Errorf("args %v env %v: quería %q\n%s%s", c.args, c.env, c.want, out, errb)
		}
	}
}

// La ayuda se compara con un archivo .golden. Si cambias la ayuda a propósito:
//
//	go test ./vigia -run Ayuda -update
func TestAyudaGolden(t *testing.T) {
	_, out, _ := correr([]string{"ayuda"}, "", nil)
	ruta := "testdata/ayuda.golden"
	if *actualizar {
		os.WriteFile(ruta, []byte(out), 0o644)
	}
	want, err := os.ReadFile(ruta)
	if err != nil {
		t.Fatal(err)
	}
	if out != string(want) {
		t.Errorf("la ayuda cambió; si es a propósito, -update\n--- quería\n%s\n--- salió\n%s", want, out)
	}
}
