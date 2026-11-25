package tarea

import (
	"context"
	"errors"
	"sync/atomic"
	"testing"
	"time"
)

func escenarioC12() (bool, bool) {
	var n atomic.Int64
	var err error
	ok1 := termina(func() {
		err = Consumir(context.Background(), finito(1, 2, 3, 4, 5, 6, 7, 8, 9, 10), 3, func(int) { n.Add(1) })
	})
	feliz := ok1 && err == nil && n.Load() == 10

	// trabajos nunca se cierra; se cancela a los 20 ms.
	ctx, cancel := context.WithCancel(context.Background())
	trabajos := make(chan int)
	go func() { time.Sleep(20 * time.Millisecond); cancel() }()
	var err2 error
	inicio := time.Now()
	ok2 := termina(func() { err2 = Consumir(ctx, trabajos, 3, func(int) {}) })
	cancela := ok2 && errors.Is(err2, context.Canceled) && time.Since(inicio) < 300*time.Millisecond
	return ok1 && ok2, feliz && cancela
}

func TestC12(t *testing.T) {
	ok, correcto := escenarioC12()
	informe(t, ok, correcto)
}
