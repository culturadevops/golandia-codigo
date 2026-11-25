package tarea

import (
	"context"
	"fmt"
	"testing"
)

func escenarioC07() (bool, bool) {
	var got []int
	ok1 := termina(func() {
		for v := range Cuadrados(context.Background(), finito(1, 2, 3)) {
			got = append(got, v)
		}
	})
	feliz := ok1 && fmt.Sprint(got) == "[1 4 9]"

	// La entrada no se acaba (y respeta ctx). Lee 2, cancela y deja de leer.
	ctx, cancel := context.WithCancel(context.Background())
	ok2 := termina(func() {
		c := Cuadrados(ctx, fuente(ctx))
		for range 2 {
			<-c
		}
		cancel()
	})
	return ok1 && ok2, feliz
}

func TestC07(t *testing.T) {
	ok, correcto := escenarioC07()
	informe(t, ok, correcto)
}
