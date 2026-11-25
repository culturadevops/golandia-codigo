package pedidos

import (
	"context"
	"testing"
	"time"
)

func TestReservarSeCortaSiElClienteCuelga(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	time.AfterFunc(50*time.Millisecond, cancel) // el cliente cuelga a los 50 ms
	inicio := time.Now()
	err := Reservar(ctx)
	dur := time.Since(inicio).Round(10 * time.Millisecond)
	if dur > 500*time.Millisecond {
		t.Fatalf("el cliente colgó a los 50ms y Reservar siguió hasta los %v (err = %v)", dur, err)
	}
	t.Logf("el cliente colgó a los 50ms -> Reservar volvió a los %v: %v", dur, err)
}
