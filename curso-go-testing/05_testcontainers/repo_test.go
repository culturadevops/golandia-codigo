//go:build integracion

package contenedores

import (
	"errors"
	"reflect"
	"strings"
	"testing"
	"time"

	"curso-testing/tienda"
)

// Ojo: estos tests NO llevan t.Parallel. Comparten una base de datos y la restauran al
// terminar; en paralelo se borrarían los datos unos a otros (lección 1, otra vez).

func pedido(cliente string) *tienda.Pedido {
	return &tienda.Pedido{Cliente: cliente, Total: 18400, Estado: "pagado", CobroID: "cob-1",
		Creado: time.Date(2026, 10, 28, 12, 0, 0, 0, time.UTC),
		Items:  []tienda.Item{{SKU: "TAZA", Cantidad: 1, Precio: 15000}, {SKU: "FILTRO", Cantidad: 2, Precio: 2000}}}
}

func TestGuardarYBuscar(t *testing.T) {
	r := repo(t)
	p := pedido("ana")
	if err := r.Guardar(t.Context(), p); err != nil {
		t.Fatal(err)
	}
	got, err := r.Buscar(t.Context(), p.ID)
	if err != nil {
		t.Fatal(err)
	}
	got.Creado = got.Creado.UTC()
	if !reflect.DeepEqual(got, p) {
		t.Errorf("leído  %+v\nguardé %+v", got, p)
	}
}

func TestActualizar(t *testing.T) {
	r := repo(t)
	p := pedido("ana")
	r.Guardar(t.Context(), p)
	p.Estado, p.Items = "enviado", p.Items[:1]
	if err := r.Guardar(t.Context(), p); err != nil {
		t.Fatal(err)
	}
	got, _ := r.Buscar(t.Context(), p.ID)
	if got.Estado != "enviado" || len(got.Items) != 1 {
		t.Errorf("después de actualizar: %+v", got)
	}
}

func TestNoExiste(t *testing.T) {
	r := repo(t)
	if _, err := r.Buscar(t.Context(), 999); !errors.Is(err, tienda.ErrNoExiste) {
		t.Errorf("err = %v", err)
	}
	p := pedido("ana")
	p.ID = 999
	if err := r.Guardar(t.Context(), p); !errors.Is(err, tienda.ErrNoExiste) {
		t.Errorf("actualizar uno que no existe: err = %v", err)
	}
}

// Cada test empieza con la base vacía: este cuenta 1 pedido aunque otros ya guardaron.
func TestAislados(t *testing.T) {
	r := repo(t)
	r.Guardar(t.Context(), pedido("ana"))
	ps, _ := r.PorCliente(t.Context(), "ana")
	if len(ps) != 1 || ps[0].ID != 1 {
		t.Errorf("quiero 1 pedido con ID 1 (base restaurada); hay %d", len(ps))
	}
}

// Lo que el fake no sabe: Postgres comprueba la cantidad (CHECK cantidad > 0) y la transacción
// deshace el pedido entero si un item falla.
func TestCantidadCeroDeshaceTodo(t *testing.T) {
	r := repo(t)
	p := pedido("ana")
	p.Items[1].Cantidad = 0
	err := r.Guardar(t.Context(), p)
	if err == nil || !strings.Contains(err.Error(), "check") {
		t.Fatalf("err = %v; quiero la violación del CHECK", err)
	}
	ps, _ := r.PorCliente(t.Context(), "ana")
	t.Logf("Guardar -> %v", err)
	t.Logf("pedidos de ana después del fallo: %d", len(ps))
	if len(ps) != 0 {
		t.Error("quedó un pedido a medias")
	}
}
