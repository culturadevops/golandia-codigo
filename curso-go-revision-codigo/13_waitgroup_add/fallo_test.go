package pedidos

import (
	"runtime"
	"testing"
	"time"
)

func TestTotalDeCienProductos(t *testing.T) {
	// Como en un pod con límite de una CPU: desde Go 1.25, GOMAXPROCS sigue ese límite.
	defer runtime.GOMAXPROCS(runtime.GOMAXPROCS(1))
	ids := make([]int, 100)
	precio := func(int) int { time.Sleep(time.Millisecond); return 10 }
	if total := TotalCarrito(ids, precio); total != 1000 {
		t.Fatalf("100 productos de 10, una CPU -> total %d: Wait volvió antes de que las goroutines hicieran su Add", total)
	}
	t.Logf("100 productos de 10, una CPU -> total 1000")
}
