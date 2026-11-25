package tarea

import (
	"strconv"
	"sync"
	"testing"
)

func escenarioC10() (bool, bool) {
	var aciertos int
	ok := termina(func() {
		c := NuevaCache()
		c.Set("k", "v")
		var wg sync.WaitGroup
		for g := range 8 {
			wg.Add(1)
			go func() {
				defer wg.Done()
				for i := range 1000 {
					c.Get("k")
					if i%100 == 0 {
						c.Set("g"+strconv.Itoa(g), "x")
					}
				}
			}()
		}
		wg.Wait()
		aciertos = c.Aciertos()
	})
	return ok, ok && aciertos == 8000
}

func TestC10(t *testing.T) {
	ok, correcto := escenarioC10()
	informe(t, ok, correcto)
}
