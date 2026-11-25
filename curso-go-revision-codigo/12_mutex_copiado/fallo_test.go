package pedidos

import (
	"sync"
	"testing"
)

func TestContadorDeCienPedidos(t *testing.T) {
	var c Contador
	var wg sync.WaitGroup
	for range 100 {
		wg.Go(c.Sumar)
	}
	wg.Wait()
	if v := c.Valor(); v != 100 {
		t.Fatalf("100 pedidos atendidos -> el contador dice %d", v)
	}
	t.Logf("100 pedidos atendidos -> el contador dice 100")
}
