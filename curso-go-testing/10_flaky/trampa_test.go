//go:build trampa

//	go test -tags trampa -run TestSKUs$ -count=100 ./10_flaky
//	go test -tags trampa -run 'TestMoneda|TestCambia' -shuffle=on ./10_flaky
//	go test -tags trampa -run TestVisitas$ -race ./10_flaky
package flaky

import (
	"slices"
	"testing"
)

// 1 · Depende del orden de un map.
func TestSKUs(t *testing.T) {
	want := []string{"TAZA", "FILTRO", "MOLINO", "CAFE"}
	if got := SKUs(carrito); !slices.Equal(got, want) {
		t.Errorf("got %v", got)
	}
}

// 2 · Depende del orden de los tests. Este va PRIMERO en el archivo y pasa...
func TestMonedaPorDefecto(t *testing.T) {
	if got := Etiqueta("10,00"); got != "10,00 MXN" {
		t.Errorf("got %q", got)
	}
}

// ...porque este, que cambia Moneda y no la restaura, corre después. Hasta que alguien los
// reordena, o -shuffle los baraja.
func TestCambiaMoneda(t *testing.T) {
	Moneda = "USD"
	if got := Etiqueta("10,00"); got != "10,00 USD" {
		t.Errorf("got %q", got)
	}
}

// 3 · Carrera de datos: casi siempre da 50... y no siempre.
func TestVisitas(t *testing.T) {
	if got := (&Visitas{}).Contar(50); got != 50 {
		t.Errorf("got %d; quiero 50", got)
	}
}
