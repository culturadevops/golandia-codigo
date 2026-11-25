package fakesmocks

import (
	"context"
	"errors"
	"testing"

	"curso-testing/tienda"
	"curso-testing/tienda/memoria"
)

// Con el FAKE se prueba el RESULTADO: qué quedó guardado y qué se cobró.
// No dice nada de qué métodos se llaman ni en qué orden.

var taza = []tienda.Item{{SKU: "TAZA", Cantidad: 1, Precio: 15000}}

func conFake() (*Servicio, *memoria.Repo, *memoria.Cobrador) {
	repo, cob := memoria.Nuevo(), &memoria.Cobrador{Rechazar: map[string]bool{"moroso": true}}
	return &Servicio{Repo: repo, Cobros: cob}, repo, cob
}

func guardado(t *testing.T, repo *memoria.Repo, id int64) *tienda.Pedido {
	t.Helper()
	p, err := repo.Buscar(context.Background(), id)
	if err != nil {
		t.Fatalf("Buscar(%d): %v", id, err)
	}
	return p
}

func TestFake_PrimerPedido(t *testing.T) {
	s, repo, _ := conFake()
	p, err := s.Crear(t.Context(), "ana", taza)
	if err != nil {
		t.Fatal(err)
	}
	if g := guardado(t, repo, p.ID); g.Total != 18400 || g.Estado != "pagado" {
		t.Errorf("guardado: total %s, estado %s; quiero 184,00 pagado", g.Total, g.Estado)
	}
}

func TestFake_SegundoPedido(t *testing.T) {
	s, repo, _ := conFake()
	s.Crear(t.Context(), "ana", taza)
	p, err := s.Crear(t.Context(), "ana", taza)
	if err != nil {
		t.Fatal(err)
	}
	if g := guardado(t, repo, p.ID); g.Total != 19900 {
		t.Errorf("total %s; quiero 199,00 (sin descuento)", g.Total)
	}
}

func TestFake_CobroRechazado(t *testing.T) {
	s, repo, _ := conFake()
	_, err := s.Crear(t.Context(), "moroso", taza)
	if !errors.Is(err, memoria.ErrRechazado) {
		t.Fatalf("err = %v", err)
	}
	ps, _ := repo.PorCliente(t.Context(), "moroso")
	for _, p := range ps {
		if p.Estado == "pagado" {
			t.Errorf("quedó un pedido pagado sin cobro: %+v", p)
		}
	}
}

func TestFake_PedidoVacio(t *testing.T) {
	s, _, cob := conFake()
	if _, err := s.Crear(t.Context(), "ana", nil); !errors.Is(err, tienda.ErrPedidoVacio) {
		t.Fatalf("err = %v", err)
	}
	if len(cob.Cobros) != 0 {
		t.Errorf("se cobró un pedido vacío: %v", cob.Cobros)
	}
}

func TestFake_EnvioGratis(t *testing.T) {
	s, repo, _ := conFake()
	s.Crear(t.Context(), "ana", taza)
	p, _ := s.Crear(t.Context(), "ana", []tienda.Item{{SKU: "MOLINO", Cantidad: 1, Precio: 50000}})
	if g := guardado(t, repo, p.ID); g.Total != 50000 {
		t.Errorf("total %s; quiero 500,00 sin envío", g.Total)
	}
}

func TestFake_QuedaPagado(t *testing.T) {
	s, repo, cob := conFake()
	p, _ := s.Crear(t.Context(), "ana", taza)
	g := guardado(t, repo, p.ID)
	if g.Estado != "pagado" || g.CobroID != "cob-ana" || len(cob.Cobros) != 1 {
		t.Errorf("guardado %+v, cobros %v", g, cob.Cobros)
	}
}
