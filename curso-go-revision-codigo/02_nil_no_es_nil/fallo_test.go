package pedidos

import "testing"

func TestValidarPedidoBueno(t *testing.T) {
	err := Validar(Pedido{Cliente: "ana", Cantidad: 3})
	if err != nil {
		t.Fatalf("pedido bueno -> err != nil es %v; tipo %T; err = %v", err != nil, err, err)
	}
	t.Logf("pedido bueno -> err = %v", err)
}
