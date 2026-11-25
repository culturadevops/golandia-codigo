package main

import "testing"

var sumidero float64

func BenchmarkTrabajo(b *testing.B) {
	for b.Loop() {
		sumidero = trabajo(200)
	}
}

// BenchmarkOtraCosa no toca el evaluador: su perfil NO es representativo de trabajo().
func BenchmarkOtraCosa(b *testing.B) {
	for b.Loop() {
		s := 0
		for i := 0; i < 20000; i++ {
			s += i % 7
		}
		sumidero = float64(s)
	}
}
