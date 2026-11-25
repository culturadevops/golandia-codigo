package fakesmocks

import (
	"context"
	"errors"
	"testing"

	"go.uber.org/mock/gomock"

	"curso-testing/tienda"
)

// Con MOCKS (gomock + mockgen) se prueba la CONVERSACIÓN: qué métodos se llaman, con qué
// argumentos y cuántas veces. Escritos como se escriben normalmente.
//
// mocks_test.go lo genera:
//   go run go.uber.org/mock/mockgen@v0.6.0 -destination=03_fakes_mocks/mocks_test.go \
//      -package=fakesmocks curso-testing/tienda Repositorio,Cobrador

func conMocks(t *testing.T) (*Servicio, *MockRepositorio, *MockCobrador) {
	ctrl := gomock.NewController(t)
	repo, cob := NewMockRepositorio(ctrl), NewMockCobrador(ctrl)
	return &Servicio{Repo: repo, Cobros: cob}, repo, cob
}

func TestMock_PrimerPedido(t *testing.T) {
	s, repo, cob := conMocks(t)
	repo.EXPECT().PorCliente(gomock.Any(), "ana").Return(nil, nil)
	cob.EXPECT().Cobrar(gomock.Any(), "ana", tienda.Monto(18400)).Return("cob-ana", nil)
	repo.EXPECT().Guardar(gomock.Any(), gomock.Any()).Return(nil)
	p, err := s.Crear(t.Context(), "ana", taza)
	if err != nil || p.Total != 18400 || p.Estado != "pagado" {
		t.Errorf("p = %+v, err = %v", p, err)
	}
}

func TestMock_SegundoPedido(t *testing.T) {
	s, repo, cob := conMocks(t)
	repo.EXPECT().PorCliente(gomock.Any(), "ana").Return([]tienda.Pedido{{ID: 1}}, nil)
	cob.EXPECT().Cobrar(gomock.Any(), "ana", tienda.Monto(19900)).Return("cob-ana", nil)
	repo.EXPECT().Guardar(gomock.Any(), gomock.Any()).Return(nil)
	if p, err := s.Crear(t.Context(), "ana", taza); err != nil || p.Total != 19900 {
		t.Errorf("p = %+v, err = %v", p, err)
	}
}

func TestMock_CobroRechazado(t *testing.T) {
	s, repo, cob := conMocks(t)
	rechazo := errors.New("tarjeta rechazada")
	repo.EXPECT().PorCliente(gomock.Any(), "moroso").Return(nil, nil)
	cob.EXPECT().Cobrar(gomock.Any(), "moroso", gomock.Any()).Return("", rechazo)
	// sin EXPECT de Guardar: si se guarda algo, el mock falla
	if _, err := s.Crear(t.Context(), "moroso", taza); !errors.Is(err, rechazo) {
		t.Errorf("err = %v", err)
	}
}

func TestMock_PedidoVacio(t *testing.T) {
	s, _, _ := conMocks(t) // ninguna llamada esperada
	if _, err := s.Crear(t.Context(), "ana", nil); !errors.Is(err, tienda.ErrPedidoVacio) {
		t.Errorf("err = %v", err)
	}
}

func TestMock_EnvioGratis(t *testing.T) {
	s, repo, cob := conMocks(t)
	repo.EXPECT().PorCliente(gomock.Any(), "ana").Return([]tienda.Pedido{{ID: 1}}, nil)
	cob.EXPECT().Cobrar(gomock.Any(), "ana", tienda.Monto(50000)).Return("cob-ana", nil)
	repo.EXPECT().Guardar(gomock.Any(), gomock.Any()).Return(nil)
	molino := []tienda.Item{{SKU: "MOLINO", Cantidad: 1, Precio: 50000}}
	if p, err := s.Crear(t.Context(), "ana", molino); err != nil || p.Total != 50000 {
		t.Errorf("p = %+v, err = %v", p, err)
	}
}

func TestMock_QuedaPagado(t *testing.T) {
	s, repo, cob := conMocks(t)
	var estado string
	repo.EXPECT().PorCliente(gomock.Any(), "ana").Return(nil, nil)
	cob.EXPECT().Cobrar(gomock.Any(), "ana", gomock.Any()).Return("cob-ana", nil)
	repo.EXPECT().Guardar(gomock.Any(), gomock.Any()).
		DoAndReturn(func(_ context.Context, p *tienda.Pedido) error {
			estado = p.Estado // lo que se guardó, en el momento de guardarlo
			return nil
		})
	s.Crear(t.Context(), "ana", taza)
	if estado != "pagado" {
		t.Errorf("se guardó con estado %q; quiero pagado", estado)
	}
}
