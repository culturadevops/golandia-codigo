package main

import (
	"strings"
	"testing"
)

// Prueba de tabla: cada fila es un caso con su nombre, su entrada y lo que espero.
func TestValidarTitulo(t *testing.T) {
	casos := []struct {
		nombre  string
		entrada string
		quiero  string
		error   bool
	}{
		{"normal", "comprar pan", "comprar pan", false},
		{"quita espacios", "  comprar pan  ", "comprar pan", false},
		{"vacío", "", "", true},
		{"solo espacios", "   ", "", true},
		{"120 letras", strings.Repeat("a", 120), strings.Repeat("a", 120), false},
		{"121 letras", strings.Repeat("a", 121), "", true},
		{"120 eñes", strings.Repeat("ñ", 120), strings.Repeat("ñ", 120), false},
	}
	for _, c := range casos {
		t.Run(c.nombre, func(t *testing.T) {
			got, err := validarTitulo(c.entrada)
			if (err != nil) != c.error {
				t.Fatalf("validarTitulo(%q): error = %v, quería error: %v", c.entrada, err, c.error)
			}
			if got != c.quiero {
				t.Errorf("validarTitulo(%q) = %q, quiero %q", c.entrada, got, c.quiero)
			}
		})
	}
}
