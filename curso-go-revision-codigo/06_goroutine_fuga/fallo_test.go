package pedidos

import (
	"runtime"
	"testing"
	"time"
)

func TestPrecioNoDejaGoroutines(t *testing.T) {
	antes := runtime.NumGoroutine()
	lento := func() int { time.Sleep(20 * time.Millisecond); return 100 }
	for range 200 {
		_, _ = Precio(lento, time.Millisecond) // el proveedor siempre llega tarde
	}
	time.Sleep(300 * time.Millisecond) // de sobra para que el proveedor conteste
	sobran := runtime.NumGoroutine() - antes
	if sobran > 5 {
		t.Fatalf("200 llamadas con plazo vencido -> %d goroutines siguen vivas, bloqueadas para siempre", sobran)
	}
	t.Logf("200 llamadas con plazo vencido -> goroutines de más: %d", sobran)
}
