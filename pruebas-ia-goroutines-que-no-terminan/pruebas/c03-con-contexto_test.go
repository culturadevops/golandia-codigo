package tarea

import (
	"context"
	"errors"
	"testing"
	"time"
)

func escenarioC03() (bool, bool) {
	var s string
	var err error
	ok1 := termina(func() {
		s, err = LlamarConContexto(context.Background(), func() (string, error) { return "hola", nil })
	})
	feliz := ok1 && err == nil && s == "hola"

	var err2 error
	inicio := time.Now()
	ok2 := termina(func() {
		for range 3 {
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Millisecond)
			_, err2 = LlamarConContexto(ctx, func() (string, error) { time.Sleep(100 * time.Millisecond); return "tarde", nil })
			cancel()
		}
	})
	corta := ok2 && errors.Is(err2, context.DeadlineExceeded) && time.Since(inicio) < 200*time.Millisecond
	return ok1 && ok2, feliz && corta
}

func TestC03(t *testing.T) {
	ok, correcto := escenarioC03()
	informe(t, ok, correcto)
}
