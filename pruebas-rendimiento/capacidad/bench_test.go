package capacidad

import "testing"

const N = 10000

func SinReservar() []int {
	var xs []int
	for i := 0; i < N; i++ {
		xs = append(xs, i) // se muda de array unas cuantas veces
	}
	return xs
}

func Reservando() []int {
	xs := make([]int, 0, N) // una sola peticion de memoria
	for i := 0; i < N; i++ {
		xs = append(xs, i)
	}
	return xs
}

var sink []int

func BenchmarkSinReservar(b *testing.B) {
	for i := 0; i < b.N; i++ {
		sink = SinReservar()
	}
}
func BenchmarkReservando(b *testing.B) {
	for i := 0; i < b.N; i++ {
		sink = Reservando()
	}
}
