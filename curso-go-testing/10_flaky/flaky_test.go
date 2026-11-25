package flaky

import (
	"slices"
	"testing"
)

// Las versiones arregladas. Aguantan -count=100, -race y -shuffle=on.

var carrito = map[string]int{"TAZA": 1, "FILTRO": 2, "MOLINO": 1, "CAFE": 3}

func TestSKUsOrdenados(t *testing.T) {
	want := []string{"CAFE", "FILTRO", "MOLINO", "TAZA"}
	if got := SKUsOrdenados(carrito); !slices.Equal(got, want) {
		t.Errorf("got %v", got)
	}
}

func TestEtiquetaUSD(t *testing.T) {
	viejo := Moneda
	t.Cleanup(func() { Moneda = viejo }) // lo deja como estaba
	Moneda = "USD"
	if got := Etiqueta("10,00"); got != "10,00 USD" {
		t.Errorf("got %q", got)
	}
}

func TestEtiquetaPorDefecto(t *testing.T) {
	if got := Etiqueta("10,00"); got != "10,00 MXN" {
		t.Errorf("got %q", got)
	}
}

func TestVisitasSeguras(t *testing.T) {
	if got := (&VisitasSeguras{}).Contar(50); got != 50 {
		t.Errorf("got %d", got)
	}
}
