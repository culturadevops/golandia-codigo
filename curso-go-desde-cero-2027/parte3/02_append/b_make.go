package main

import (
	"fmt"
	"testing"
)

func sinReservar() []int {
	var s []int
	for i := 0; i < 100000; i++ {
		s = append(s, i)
	}
	return s
}

func reservando() []int {
	s := make([]int, 0, 100000)
	for i := 0; i < 100000; i++ {
		s = append(s, i)
	}
	return s
}

func main() {
	for _, c := range []struct {
		nombre string
		f      func() []int
	}{{"sin reservar", sinReservar}, {"make con cap", reservando}} {
		r := testing.Benchmark(func(b *testing.B) {
			b.ReportAllocs()
			for i := 0; i < b.N; i++ {
				c.f()
			}
		})
		fmt.Printf("%-13s %8d ns/op  %3d reservas/op\n", c.nombre, r.NsPerOp(), r.AllocsPerOp())
	}
}
