package reintentos

import (
	"testing"
	"testing/synctest"
	"time"
)

func TestDebounce(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		t0 := time.Now()
		var veces int
		var cuando time.Duration
		buscar := Debounce(300*time.Millisecond, func() { veces++; cuando = time.Since(t0) })
		for i := 0; i < 5; i++ { // cinco teclas, una cada 100 ms
			buscar()
			time.Sleep(100 * time.Millisecond)
		}
		synctest.Wait() // nada que hacer todavía: el temporizador sigue pendiente
		t.Logf("tras teclear: %d búsquedas", veces)
		time.Sleep(time.Second)
		t.Logf("al final: %d búsqueda, a los %v", veces, cuando)
		if veces != 1 || cuando != 700*time.Millisecond {
			t.Fatal("no")
		}
	})
}
