package main

import "testing"

func TestDescuento(t *testing.T) {
	casos := []struct {
		nombre     string
		precio     int
		porcentaje int
		quiero     int
	}{
		{"sin descuento", 1000, 0, 1000},
		{"diez por ciento", 1000, 10, 900},
		{"todo gratis", 1000, 100, 0},
		{"más de cien", 1000, 150, 0},
		{"porcentaje negativo", 1000, -20, 1000},
	}
	for _, c := range casos {
		t.Run(c.nombre, func(t *testing.T) {
			if got := Descuento(c.precio, c.porcentaje); got != c.quiero {
				t.Errorf("Descuento(%d, %d) = %d, quiero %d", c.precio, c.porcentaje, got, c.quiero)
			}
		})
	}
}
