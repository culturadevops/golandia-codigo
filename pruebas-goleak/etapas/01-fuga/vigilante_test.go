package fuga

import (
	"context"
	"testing"
	"time"

	"go.uber.org/goleak"
)

// ── 1. El test que PASA con la fuga dentro ───────────────────────────────
// Sin goleak, este test es verde. Comprueba lo que dice comprobar, y lo
// comprueba bien. La goroutine que queda viva no le importa a nadie.
func TestConFuga_SinGoleak(t *testing.T) {
	v := &VigilanteConFuga{}
	v.Arrancar(10 * time.Millisecond)
	time.Sleep(50 * time.Millisecond)
	if v.Vistas() == 0 {
		t.Fatal("deberia haber contado algo")
	}
}

// ── 2. El MISMO test, con goleak al final ────────────────────────────────
// Una línea. Y ahora falla, diciendo dónde nació la goroutine que no murió.
func TestConFuga_ConGoleak(t *testing.T) {
	defer goleak.VerifyNone(t)

	v := &VigilanteConFuga{}
	v.Arrancar(10 * time.Millisecond)
	time.Sleep(50 * time.Millisecond)
	if v.Vistas() == 0 {
		t.Fatal("deberia haber contado algo")
	}
}

// ── 3. La versión sana: mismo test, y goleak no se queja ─────────────────
func TestSano_ConGoleak(t *testing.T) {
	defer goleak.VerifyNone(t)

	ctx, cancel := context.WithCancel(context.Background())
	v := &VigilanteSano{}
	v.Arrancar(ctx, 10*time.Millisecond)
	time.Sleep(50 * time.Millisecond)
	if v.Vistas() == 0 {
		t.Fatal("deberia haber contado algo")
	}
	cancel()
	// ⚠️ Sin esta espera, goleak puede mirar ANTES de que la goroutine haya
	// terminado de salir y dar un falso positivo. Es la queja número uno.
	time.Sleep(20 * time.Millisecond)
}
