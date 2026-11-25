package tarea

import (
	"context"
	"fmt"
	"testing"
)

func escenarioC06() (bool, bool) {
	var got []int
	ok1 := termina(func() {
		for v := range Generar(context.Background(), 5) {
			got = append(got, v)
		}
	})
	feliz := ok1 && fmt.Sprint(got) == "[0 1 2 3 4]"

	// Lee 3, cancela y deja de leer.
	ctx, cancel := context.WithCancel(context.Background())
	ok2 := termina(func() {
		c := Generar(ctx, 1_000_000)
		for range 3 {
			<-c
		}
		cancel()
	})
	return ok1 && ok2, feliz
}

func TestC06(t *testing.T) {
	ok, correcto := escenarioC06()
	informe(t, ok, correcto)
}
