package etiquetas

import (
	"testing"

	"curso-testing/tienda"
	"curso-testing/tienda/memoria"
)

// El test rápido: contra el fake, siempre.
func TestPagadosDe(t *testing.T) {
	r := memoria.Nuevo()
	for _, e := range []string{"pagado", "rechazado", "pagado"} {
		r.Guardar(t.Context(), &tienda.Pedido{Cliente: "ana", Estado: e})
	}
	if n, _ := PagadosDe(t.Context(), r, "ana"); n != 2 {
		t.Errorf("n = %d; quiero 2", n)
	}
}
