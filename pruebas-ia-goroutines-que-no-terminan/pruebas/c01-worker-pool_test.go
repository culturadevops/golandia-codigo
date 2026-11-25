package tarea

import (
	"context"
	"fmt"
	"testing"
	"time"
)

func escenarioC01() (bool, bool) {
	var res []int
	var err error
	ok1 := termina(func() {
		res, err = ProcesarConWorkers(context.Background(), []int{1, 2, 3, 4, 5}, 2, func(x int) int { return x * 10 })
	})
	feliz := ok1 && err == nil && fmt.Sprint(res) == "[10 20 30 40 50]"

	ctx, cancel := context.WithCancel(context.Background())
	trabajos := make([]int, 400)
	go func() { time.Sleep(20 * time.Millisecond); cancel() }()
	var err2 error
	inicio := time.Now()
	ok2 := termina(func() {
		_, err2 = ProcesarConWorkers(ctx, trabajos, 4, func(x int) int { time.Sleep(10 * time.Millisecond); return x })
	})
	cancela := ok2 && err2 != nil && time.Since(inicio) < 300*time.Millisecond
	return ok1 && ok2, feliz && cancela
}

func TestC01(t *testing.T) {
	ok, correcto := escenarioC01()
	informe(t, ok, correcto)
}
