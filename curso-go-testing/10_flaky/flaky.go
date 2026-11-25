// Lección 10 · Tests flaky: detectarlos con -count, -race y -shuffle, y arreglarlos.
package flaky

import (
	"slices"
	"sync"
)

// SKUs devuelve los productos de un carrito. Recorre un map: el orden cambia en cada vuelta.
func SKUs(carrito map[string]int) []string {
	var out []string
	for sku := range carrito {
		out = append(out, sku)
	}
	return out
}

// SKUsOrdenados es el arreglo cuando el orden importa (o el test lo compara sin orden).
func SKUsOrdenados(carrito map[string]int) []string {
	out := SKUs(carrito)
	slices.Sort(out)
	return out
}

// Moneda es estado global del paquete. Un test que lo cambia y no lo deja como estaba
// rompe a los que corren después... según el orden.
var Moneda = "MXN"

func Etiqueta(precio string) string { return precio + " " + Moneda }

// Visitas cuenta desde varias goroutines. Sin mutex: es una carrera de datos.
type Visitas struct{ n int }

func (v *Visitas) Contar(goroutines int) int {
	var wg sync.WaitGroup
	for range goroutines {
		wg.Go(func() { v.n++ })
	}
	wg.Wait()
	return v.n
}

// VisitasSeguras es el arreglo.
type VisitasSeguras struct {
	mu sync.Mutex
	n  int
}

func (v *VisitasSeguras) Contar(goroutines int) int {
	var wg sync.WaitGroup
	for range goroutines {
		wg.Go(func() {
			v.mu.Lock()
			v.n++
			v.mu.Unlock()
		})
	}
	wg.Wait()
	return v.n
}
