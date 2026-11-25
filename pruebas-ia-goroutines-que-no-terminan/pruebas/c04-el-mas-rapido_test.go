package tarea

import (
	"testing"
	"time"
)

func lenta(d time.Duration, s string) func() string {
	return func() string { time.Sleep(d); return s }
}

func escenarioC04() (bool, bool) {
	var r string
	inicio := time.Now()
	ok := termina(func() {
		r = ElMasRapido([]func() string{lenta(80*time.Millisecond, "c"), lenta(10*time.Millisecond, "a"), lenta(50*time.Millisecond, "b")})
	})
	return ok, ok && r == "a" && time.Since(inicio) < 45*time.Millisecond
}

func TestC04(t *testing.T) {
	ok, correcto := escenarioC04()
	informe(t, ok, correcto)
}
