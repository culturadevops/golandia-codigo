package pp

import (
	"strings"
	"testing"
)

var base = strings.Fields(strings.Repeat("El Rapido Zorro Marron Salta Sobre El Perro Perezoso ", 4000))

func lento(palabras []string) map[string]int {
	cuenta := map[string]int{}
	for _, p := range palabras {
		clave := ""
		for _, r := range strings.ToLower(p) {
			if r >= 'a' && r <= 'z' {
				clave += string(r)
			}
		}
		cuenta[clave]++
	}
	return cuenta
}

func rapido(palabras []string) map[string]int {
	cuenta := map[string]int{}
	var b strings.Builder
	for _, p := range palabras {
		b.Reset()
		b.Grow(len(p))
		for i := 0; i < len(p); i++ {
			c := p[i]
			if c >= 'A' && c <= 'Z' {
				c += 'a' - 'A'
			}
			if c >= 'a' && c <= 'z' {
				b.WriteByte(c)
			}
		}
		cuenta[b.String()]++
	}
	return cuenta
}

var sink map[string]int

func BenchmarkAntes(b *testing.B) {
	for i := 0; i < b.N; i++ {
		sink = lento(base)
	}
}
func BenchmarkDespues(b *testing.B) {
	for i := 0; i < b.N; i++ {
		sink = rapido(base)
	}
}
