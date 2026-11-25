package tarea

import (
	"testing"
	"time"
)

func escenarioC02() (bool, bool) {
	var v int
	var err error
	ok1 := termina(func() { v, err = ConTimeout(200*time.Millisecond, func() int { return 3 }) })
	feliz := ok1 && err == nil && v == 3

	var err2 error
	inicio := time.Now()
	ok2 := termina(func() {
		for range 3 {
			_, err2 = ConTimeout(5*time.Millisecond, func() int { time.Sleep(100 * time.Millisecond); return 7 })
		}
	})
	corta := ok2 && err2 != nil && time.Since(inicio) < 200*time.Millisecond
	return ok1 && ok2, feliz && corta
}

func TestC02(t *testing.T) {
	ok, correcto := escenarioC02()
	informe(t, ok, correcto)
}
