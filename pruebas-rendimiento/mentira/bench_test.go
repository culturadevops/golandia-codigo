package mentira

import "testing"

type Pedido struct {
	ID    int64
	Notas [256]byte
}

// Funcion normal: pequena y por tanto INCRUSTABLE. Es el caso realista.
func total(p Pedido) int64 { return p.ID * 2 }

// EL BANCO QUE MIENTE: no se usa el resultado. El compilador incrusta la funcion,
// ve que nadie mira el valor, y borra el trabajo entero. El numero no mide nada.
func BenchmarkMiente(b *testing.B) {
	p := Pedido{ID: 42}
	for i := 0; i < b.N; i++ {
		total(p)
	}
}

// EL BANCO HONESTO: el resultado va a una variable de paquete, que el compilador
// no puede demostrar que sobre. Ahora si se ejecuta.
var sumidero int64

func BenchmarkHonesto(b *testing.B) {
	p := Pedido{ID: 42}
	for i := 0; i < b.N; i++ {
		sumidero = total(p)
	}
}
