package regla

import "testing"

// El caso que todo el mundo ha visto en una revision de codigo: «saca el len() del bucle,
// que se recalcula en cada vuelta». Suena a mejora. Vamos a medirlo.

func SumaIngenua(xs []int) int {
	t := 0
	for i := 0; i < len(xs); i++ { // len() en la condicion
		t += xs[i]
	}
	return t
}

func SumaOptimizada(xs []int) int {
	t := 0
	n := len(xs) // «lo saco fuera para que no se recalcule»
	for i := 0; i < n; i++ {
		t += xs[i]
	}
	return t
}

var datos = make([]int, 100000)
var sink int

func BenchmarkIngenua(b *testing.B) {
	for i := 0; i < b.N; i++ {
		sink = SumaIngenua(datos)
	}
}

func BenchmarkOptimizada(b *testing.B) {
	for i := 0; i < b.N; i++ {
		sink = SumaOptimizada(datos)
	}
}
