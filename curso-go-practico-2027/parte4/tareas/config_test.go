package main

import (
	"strings"
	"testing"
	"time"
)

// entorno devuelve un getenv falso a partir de un map.
func entorno(vars map[string]string) func(string) string {
	return func(k string) string { return vars[k] }
}

func TestLeerConfigValoresPorDefecto(t *testing.T) {
	c, err := LeerConfig(entorno(map[string]string{"DATABASE_URL": "postgres://x"}))
	if err != nil {
		t.Fatal(err)
	}
	if c.Puerto != 8080 || c.TimeoutBD != 2*time.Second || c.MaxConexiones != 4 {
		t.Errorf("valores por defecto: %+v", c)
	}
}

func TestLeerConfigJuntaLosErrores(t *testing.T) {
	_, err := LeerConfig(entorno(map[string]string{"TAREAS_PUERTO": "ochenta", "TAREAS_TIMEOUT_BD": "2"}))
	if err == nil {
		t.Fatal("quería error")
	}
	for _, trozo := range []string{"TAREAS_PUERTO", "DATABASE_URL", "TAREAS_TIMEOUT_BD"} {
		if !strings.Contains(err.Error(), trozo) {
			t.Errorf("al error le falta %s:\n%v", trozo, err)
		}
	}
}
