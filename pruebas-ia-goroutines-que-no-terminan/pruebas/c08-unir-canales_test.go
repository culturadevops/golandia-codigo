package tarea

import (
	"context"
	"testing"
)

func escenarioC08() (bool, bool) {
	suma := 0
	ok1 := termina(func() {
		for v := range Unir(context.Background(), finito(1, 2), finito(3), finito(4, 5)) {
			suma += v
		}
	})
	feliz := ok1 && suma == 15

	// Tres entradas que no se acaban (y respetan ctx). Lee 5, cancela y deja de leer.
	ctx, cancel := context.WithCancel(context.Background())
	ok2 := termina(func() {
		c := Unir(ctx, fuente(ctx), fuente(ctx), fuente(ctx))
		for range 5 {
			<-c
		}
		cancel()
	})
	return ok1 && ok2, feliz
}

func TestC08(t *testing.T) {
	ok, correcto := escenarioC08()
	informe(t, ok, correcto)
}
