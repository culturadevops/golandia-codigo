package reintentos

import (
	"context"
	"errors"
	"testing"
	"testing/synctest"
	"time"

	"golang.org/x/time/rate"
)

var fallo = errors.New("fallo")

func TestAciertaAlSegundo(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		t0, n := time.Now(), 0
		intentos, err := Hacer(context.Background(), func() error {
			if n++; n < 2 {
				return fallo
			}
			return nil
		})
		t.Logf("intentos %d · err %v · reloj falso %v", intentos, err, time.Since(t0))
		if intentos != 2 || err != nil || time.Since(t0) != time.Second {
			t.Fatal("no")
		}
	})
}

func TestSeRinde(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		t0 := time.Now()
		intentos, err := Hacer(context.Background(), func() error { return fallo })
		t.Logf("intentos %d · err %v · reloj falso %v", intentos, err, time.Since(t0))
		if !errors.Is(err, ErrRendido) || time.Since(t0) != 15*time.Second {
			t.Fatal("no")
		}
	})
}

func TestCancelado(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		t0 := time.Now()
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		intentos, err := Hacer(ctx, func() error { return fallo })
		t.Logf("intentos %d · err %v · reloj falso %v", intentos, err, time.Since(t0))
		if !errors.Is(err, context.DeadlineExceeded) || time.Since(t0) != 5*time.Second {
			t.Fatal("no")
		}
	})
}

// Una librería que no sabe nada de synctest: el limitador de x/time/rate usa time.Now.
func TestLimitador(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		lim := rate.NewLimiter(2, 1) // 2 por segundo, ráfaga de 1
		t0 := time.Now()
		for i := 0; i < 10; i++ {
			if err := lim.Wait(context.Background()); err != nil {
				t.Fatal(err)
			}
		}
		t.Logf("10 peticiones a 2 por segundo · reloj falso %v", time.Since(t0))
	})
}

// Una tarea cada hora durante un año.
func TestUnAno(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		tk := time.NewTicker(time.Hour)
		defer tk.Stop()
		n := 0
		for n < 365*24 {
			<-tk.C
			n++
		}
		t.Logf("%d ticks · ahora es %v", n, time.Now().UTC().Format("2006-01-02 15:04"))
	})
}
