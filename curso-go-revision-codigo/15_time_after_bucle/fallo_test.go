package pedidos

import (
	"testing"
	"time"
)

func TestRecogerRespetaElPlazo(t *testing.T) {
	ch := make(chan string)
	go func() { // un almacén que confirma cada 10 ms, durante 3 s
		for fin := time.Now().Add(3 * time.Second); time.Now().Before(fin); {
			select {
			case ch <- "ok":
			case <-time.After(10 * time.Millisecond):
			}
			time.Sleep(10 * time.Millisecond)
		}
	}()
	inicio := time.Now()
	r := Recoger(ch, 100*time.Millisecond)
	dur := time.Since(inicio).Round(10 * time.Millisecond)
	if dur > 500*time.Millisecond {
		t.Fatalf("plazo de 100ms -> Recoger tardó %v (%d confirmaciones): cada mensaje reinicia el plazo", dur, len(r))
	}
	t.Logf("plazo de 100ms -> Recoger volvió a los %v con %d confirmaciones", dur, len(r))
}
