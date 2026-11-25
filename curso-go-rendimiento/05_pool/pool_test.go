// Lección 5: cuándo NO sirve sync.Pool.
package pool

import (
	"runtime"
	"sync"
	"testing"
)

// Un objeto pequeño, de los que el asignador de Go crea muy rápido.
type acumulado struct {
	total, errores int
	p50, p95       float64
}

var sumidero *acumulado

var pequenos = sync.Pool{New: func() any { return new(acumulado) }}

var sumaTotal int

func BenchmarkPequeno(b *testing.B) {
	// Si el objeto no escapa, vive en la pila y no cuesta nada: no hay qué ahorrar.
	b.Run("con=pila", func(b *testing.B) {
		for b.Loop() {
			var a acumulado
			a.total++
			sumaTotal += a.total
		}
	})
	b.Run("con=new", func(b *testing.B) {
		for b.Loop() {
			a := new(acumulado)
			a.total++
			sumidero = a
		}
	})
	b.Run("con=Pool", func(b *testing.B) {
		for b.Loop() {
			a := pequenos.Get().(*acumulado)
			*a = acumulado{}
			a.total++
			sumidero = a
			pequenos.Put(a)
		}
	})
}

// Un búfer de 64 KB: aquí el pool sí ahorra.
var grandes = sync.Pool{New: func() any { b := make([]byte, 64<<10); return &b }}

var sumideroB []byte

func BenchmarkGrande(b *testing.B) {
	b.Run("con=make", func(b *testing.B) {
		for b.Loop() {
			s := make([]byte, 64<<10)
			s[0] = 1
			sumideroB = s
		}
	})
	b.Run("con=Pool", func(b *testing.B) {
		for b.Loop() {
			p := grandes.Get().(*[]byte)
			(*p)[0] = 1
			sumideroB = *p
			grandes.Put(p)
		}
	})
}

// El pool se vacía con el recolector: lo que guardas no está garantizado.
func TestSobreviveAlGC(t *testing.T) {
	var p sync.Pool
	marca := new(acumulado)
	for gcs := range 3 {
		p.Put(marca)
		for range gcs {
			runtime.GC()
		}
		x := p.Get()
		t.Logf("Put, %d GC, Get -> ¿el mismo objeto? %v", gcs, x == marca)
	}
}
