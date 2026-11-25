package p03

import "testing"

// Un struct realista de API: no es enorme, pero tampoco dos enteros.
type Pedido struct {
	ID           int64
	Cliente      string
	Direccion    string
	Lineas       [16]int64
	Total        float64
	Notas        string
	Etiquetas    [8]string
	Observacion  [128]byte
}

//go:noinline
func (p Pedido) TotalValor() float64 { return p.Total }

//go:noinline
func (p *Pedido) TotalPuntero() float64 { return p.Total }

var sink float64

func BenchmarkReceptorValor(b *testing.B) {
	p := Pedido{Total: 42}
	for i := 0; i < b.N; i++ {
		sink = p.TotalValor()
	}
}

func BenchmarkReceptorPuntero(b *testing.B) {
	p := Pedido{Total: 42}
	for i := 0; i < b.N; i++ {
		sink = p.TotalPuntero()
	}
}
