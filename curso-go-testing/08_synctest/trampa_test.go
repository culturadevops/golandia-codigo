//go:build trampa

//	go test -tags trampa -v ./08_synctest
package relojes

import (
	"context"
	"errors"
	"testing"
	"testing/synctest"
	"time"
)

// El mismo test de reintentos SIN synctest: espera de verdad.
func TestReintentosRelojDeVerdad(t *testing.T) {
	p := &pasarela{codigos: []int{503, 503, 503, 201}}
	t0 := time.Now()
	if _, err := cliente(p).Cobrar(t.Context(), "ana", 1250); err != nil {
		t.Fatal(err)
	}
	t.Logf("reloj de verdad: %.2f s", time.Since(t0).Seconds())
}

// La versión con time.Sleep: el context vence a los 2,5 s pero el Sleep de la segunda espera
// sigue hasta los 3 s. Con reloj de verdad, «tardó un poco más» no lo ve nadie; aquí es exacto.
func TestSleepNoSeEntera(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		ctx, cancel := context.WithTimeout(t.Context(), 2500*time.Millisecond)
		defer cancel()
		t0 := time.Now()
		err := ReintentarConSleep(ctx, 4, time.Second, func() error { return errors.New("503") })
		t.Logf("ReintentarConSleep: sale a los %v con %v", time.Since(t0), err)
		if time.Since(t0) != 2500*time.Millisecond {
			t.Errorf("quiero salir a los 2.5s, cuando vence el context")
		}
	})
}
