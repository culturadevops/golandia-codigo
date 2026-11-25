// Go 1.27: rutinas de asignación especializadas por tamaño para objetos pequeños (<80 bytes).
// Las notas: «reducing the cost of some small (<80 byte) memory allocations by up to 30%».
// 128 bytes queda fuera del rango: sirve de control.
package demo

import "testing"

type obj16 struct{ a, b int64 }
type obj32 struct{ a, b, c, d int64 }
type obj64 struct{ a [8]int64 }
type obj128 struct{ a [16]int64 }

var sink any

func BenchmarkAlloc16(b *testing.B) {
	for b.Loop() {
		sink = &obj16{}
	}
}

func BenchmarkAlloc32(b *testing.B) {
	for b.Loop() {
		sink = &obj32{}
	}
}

func BenchmarkAlloc64(b *testing.B) {
	for b.Loop() {
		sink = &obj64{}
	}
}

func BenchmarkAlloc128(b *testing.B) {
	for b.Loop() {
		sink = &obj128{}
	}
}
