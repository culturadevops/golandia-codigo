package texto

import (
	"bytes"
	"strings"
	"testing"
)

const N = 2000

func ConMas() string {
	s := ""
	for i := 0; i < N; i++ {
		s += "linea de registro\n" // cada vuelta crea una cadena NUEVA
	}
	return s
}

func ConBuilder() string {
	var b strings.Builder
	for i := 0; i < N; i++ {
		b.WriteString("linea de registro\n")
	}
	return b.String()
}

func ConBuilderReservado() string {
	var b strings.Builder
	b.Grow(N * len("linea de registro\n")) // se sabe cuanto va a ocupar: se pide de una vez
	for i := 0; i < N; i++ {
		b.WriteString("linea de registro\n")
	}
	return b.String()
}

func ConBuffer() string {
	var b bytes.Buffer
	for i := 0; i < N; i++ {
		b.WriteString("linea de registro\n")
	}
	return b.String()
}

var s string

func BenchmarkConcatMas(b *testing.B) {
	for i := 0; i < b.N; i++ {
		s = ConMas()
	}
}
func BenchmarkConcatBuilder(b *testing.B) {
	for i := 0; i < b.N; i++ {
		s = ConBuilder()
	}
}
func BenchmarkConcatBuilderReservado(b *testing.B) {
	for i := 0; i < b.N; i++ {
		s = ConBuilderReservado()
	}
}
func BenchmarkConcatBuffer(b *testing.B) {
	for i := 0; i < b.N; i++ {
		s = ConBuffer()
	}
}
