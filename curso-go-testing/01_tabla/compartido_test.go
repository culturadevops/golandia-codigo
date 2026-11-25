//go:build trampa

// La trampa de la lección 1: un fake COMPARTIDO entre subtests paralelos.
// Se reinicia el contador al empezar cada subtest... y funciona mientras corren de uno en uno.
//
//	go test -tags trampa -run TestCompartido -count=50 ./01_tabla
//	go test -tags trampa -run TestCompartido -race ./01_tabla
package tabla

import (
	"context"
	"testing"
	"time"

	"curso-testing/tienda"
	"curso-testing/tienda/memoria"
)

// contador es un fake de la pasarela que cuenta cuántas veces se le cobró.
type contador struct{ llamadas int }

func (c *contador) Cobrar(context.Context, string, tienda.Monto) (string, error) {
	time.Sleep(time.Millisecond) // lo que tarda la pasarela en contestar, por poco que sea
	c.llamadas++
	return "cob", nil
}

var clientes = []string{"ana", "beto", "caro", "dani"}

func TestCompartido(t *testing.T) {
	cobros := &contador{} // UNO para todos los subtests
	for _, cl := range clientes {
		t.Run(cl, func(t *testing.T) {
			t.Parallel()
			cobros.llamadas = 0 // «lo reinicio en cada subtest»
			s := &tienda.Servicio{Repo: memoria.Nuevo(), Cobros: cobros}
			s.Crear(t.Context(), cl, []tienda.Item{{SKU: "TAZA", Cantidad: 1, Precio: 100}})
			if cobros.llamadas != 1 {
				t.Errorf("se cobró %d veces; quiero 1", cobros.llamadas)
			}
		})
	}
}

// El arreglo: cada subtest construye su propio fake. Nada compartido, nada que reiniciar.
func TestPorSubtest(t *testing.T) {
	for _, cl := range clientes {
		t.Run(cl, func(t *testing.T) {
			t.Parallel()
			cobros := &contador{}
			s := &tienda.Servicio{Repo: memoria.Nuevo(), Cobros: cobros}
			s.Crear(t.Context(), cl, []tienda.Item{{SKU: "TAZA", Cantidad: 1, Precio: 100}})
			if cobros.llamadas != 1 {
				t.Errorf("se cobró %d veces; quiero 1", cobros.llamadas)
			}
		})
	}
}
