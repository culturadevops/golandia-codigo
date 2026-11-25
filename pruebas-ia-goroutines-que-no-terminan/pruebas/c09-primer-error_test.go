package tarea

import (
	"errors"
	"testing"
	"time"
)

func escenarioC09() (bool, bool) {
	var err error
	ok1 := termina(func() { err = ProcesarTodos([]string{"a", "b", "c"}, func(string) error { return nil }) })
	feliz := ok1 && err == nil

	// Uno falla al instante; los otros 9 tardan 50 ms y también fallan.
	items := []string{"mal", "1", "2", "3", "4", "5", "6", "7", "8", "9"}
	var err2 error
	inicio := time.Now()
	ok2 := termina(func() {
		err2 = ProcesarTodos(items, func(s string) error {
			if s == "mal" {
				return errors.New("falla " + s)
			}
			time.Sleep(50 * time.Millisecond)
			return errors.New("tarde " + s)
		})
	})
	pronto := ok2 && err2 != nil && err2.Error() == "falla mal" && time.Since(inicio) < 40*time.Millisecond
	return ok1 && ok2, feliz && pronto
}

func TestC09(t *testing.T) {
	ok, correcto := escenarioC09()
	informe(t, ok, correcto)
}
