// Lección 1: el mismo código medido de dos formas.
package benchmarks

import "testing"

// esError decide si una petición cuenta como error: 5xx.
// Es pequeña: el compilador la inserta (inline) en quien la llama.
func esError(estado int) bool {
	return estado >= 500 && estado <= 599
}

// Una constante, como en tantos benchmarks: el compilador conoce la entrada.
const estado = 503

// La forma de siempre. El resultado no se usa, así que el compilador puede quitar la llamada.
func BenchmarkEsError_bN(b *testing.B) {
	for i := 0; i < b.N; i++ {
		esError(estado)
	}
}

// Con b.Loop (Go 1.24+), lo que hay dentro del bucle no se elimina aunque no uses el resultado.
func BenchmarkEsError_Loop(b *testing.B) {
	for b.Loop() {
		esError(estado)
	}
}
