package pedidos

import "testing"

func TestDecodificarRechazaLoIlegible(t *testing.T) {
	cuerpo := []byte(`{"cliente":"ana","cantidad":"tres"}`)
	p, err := Decodificar(cuerpo)
	if err == nil {
		t.Fatalf("cantidad \"tres\" -> %+v, err = nil: el pedido entra con cantidad 0", p)
	}
	t.Logf("cantidad \"tres\" -> %v", err)
}
