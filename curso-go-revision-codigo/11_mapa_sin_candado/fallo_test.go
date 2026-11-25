package pedidos

import (
	"sync"
	"testing"
)

func TestVenderDesdeOchoHandlers(t *testing.T) {
	inv := NuevoInventario()
	var wg sync.WaitGroup
	for range 8 {
		wg.Go(func() {
			for range 1000 {
				inv.Vender("taza")
			}
		})
	}
	wg.Wait()
	if s := inv.Stock("taza"); s != -8000 {
		t.Fatalf("8 handlers x 1000 ventas -> stock %d; esperaba -8000", s)
	}
	t.Logf("8 handlers x 1000 ventas -> stock -8000")
}
