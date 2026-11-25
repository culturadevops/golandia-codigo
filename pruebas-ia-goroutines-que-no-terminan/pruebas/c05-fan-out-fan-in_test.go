package tarea

import (
	"context"
	"testing"
	"time"
)

func escenarioC05() (bool, bool) {
	suma := 0
	ok1 := termina(func() {
		for v := range FanOutFanIn(context.Background(), []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}, 3, func(x int) int { return x * x }) {
			suma += v
		}
	})
	feliz := ok1 && suma == 385

	// Lee 3 resultados, cancela y deja de leer.
	ctx, cancel := context.WithCancel(context.Background())
	entrada := make([]int, 1000)
	ok2 := termina(func() {
		out := FanOutFanIn(ctx, entrada, 4, func(x int) int { time.Sleep(time.Millisecond); return x })
		for range 3 {
			<-out
		}
		cancel()
	})
	return ok1 && ok2, feliz
}

func TestC05(t *testing.T) {
	ok, correcto := escenarioC05()
	informe(t, ok, correcto)
}
