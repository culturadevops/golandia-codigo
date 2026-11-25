package pedidos

import (
	"runtime"
	"strconv"
	"testing"
)

func pedidos(n int) []Pedido {
	ps := make([]Pedido, n)
	for i := range ps {
		ps[i] = Pedido{ID: 1000 + i, Cliente: "cliente-" + strconv.Itoa(i), Total: 250}
	}
	return ps
}

func TestCSVNoReservaDeMas(t *testing.T) {
	ps := pedidos(5000)
	var antes, despues runtime.MemStats
	runtime.GC()
	runtime.ReadMemStats(&antes)
	csv := CSV(ps)
	runtime.ReadMemStats(&despues)
	kb := (despues.TotalAlloc - antes.TotalAlloc) / 1024
	if limite := uint64(10 * len(csv) / 1024); kb > limite { // 10 veces el resultado, de sobra
		t.Fatalf("CSV de 5000 pedidos (%d KB) -> reservó %d KB de memoria", len(csv)/1024, kb)
	}
	t.Logf("CSV de 5000 pedidos (%d KB) -> reservó %d KB de memoria", len(csv)/1024, kb)
}

func BenchmarkCSV(b *testing.B) {
	ps := pedidos(5000)
	for b.Loop() {
		CSV(ps)
	}
}
