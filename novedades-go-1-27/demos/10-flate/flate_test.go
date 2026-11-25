// Go 1.27: compress/flate comprime más rápido (y la salida exacta puede cambiar).
package demo

import (
	"bytes"
	"compress/gzip"
	"fmt"
	"math/rand/v2"
	"strings"
	"testing"
)

// datos: ~4 MB de "logs" deterministas (misma semilla en las dos versiones).
var datos = func() []byte {
	r := rand.New(rand.NewPCG(27, 26))
	niveles := []string{"INFO", "WARN", "ERROR", "DEBUG"}
	rutas := []string{"/api/usuarios", "/api/pedidos", "/salud", "/api/pagos", "/login"}
	var b strings.Builder
	for b.Len() < 4<<20 {
		fmt.Fprintf(&b, "2026-10-29T12:%02d:%02d %s %s %d %dms id=%08x\n",
			r.IntN(60), r.IntN(60), niveles[r.IntN(4)], rutas[r.IntN(5)],
			[]int{200, 201, 404, 500}[r.IntN(4)], r.IntN(900), r.Uint32())
	}
	return []byte(b.String())
}()

func comprimir(nivel int) []byte {
	var out bytes.Buffer
	w, _ := gzip.NewWriterLevel(&out, nivel)
	w.Write(datos)
	w.Close()
	return out.Bytes()
}

func TestTamano(t *testing.T) {
	for _, n := range []int{gzip.BestSpeed, gzip.DefaultCompression, gzip.BestCompression} {
		t.Logf("nivel %d: %d -> %d bytes", n, len(datos), len(comprimir(n)))
	}
}

func BenchmarkGzipDefault(b *testing.B) {
	b.SetBytes(int64(len(datos)))
	for b.Loop() {
		comprimir(gzip.DefaultCompression)
	}
}

func BenchmarkGzipBestSpeed(b *testing.B) {
	b.SetBytes(int64(len(datos)))
	for b.Loop() {
		comprimir(gzip.BestSpeed)
	}
}
