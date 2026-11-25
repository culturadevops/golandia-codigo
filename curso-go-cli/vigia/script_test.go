package main

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/rogpeppe/go-internal/testscript"
)

// TestMain convierte este binario de test en «vigia»: los scripts de testdata/script
// lo ejecutan como un programa de verdad, con su proceso, sus pipes y su código de salida.
func TestMain(m *testing.M) {
	testscript.Main(m, map[string]func(){"vigia": main})
}

func TestScripts(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/ok", func(w http.ResponseWriter, r *http.Request) {})
	mux.HandleFunc("/roto", func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(500) })
	srv := httptest.NewServer(mux)
	// Cleanup y no defer: testscript corre cada script como subtest en paralelo, y esos
	// subtests siguen vivos cuando esta función ya volvió. Con defer, el servidor se
	// cierra antes de que empiecen y todo da «connection refused».
	t.Cleanup(srv.Close)

	testscript.Run(t, testscript.Params{
		Dir: "testdata/script",
		Setup: func(env *testscript.Env) error {
			env.Setenv("SRV", srv.URL)
			return nil
		},
	})
}
