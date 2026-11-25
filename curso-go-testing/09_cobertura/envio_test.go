package cobertura

import "testing"

// Dos casos, uno por rama: 100 % de cobertura. Y el fallo de 500,00 sigue ahí.
//
//	go test -cover ./09_cobertura
//	go test -coverprofile=cover.out ./09_cobertura && go tool cover -func=cover.out
func TestEnvio(t *testing.T) {
	if got := Envio(10000); got != Costo {
		t.Errorf("Envio(100,00) = %d", got)
	}
	if got := Envio(90000); got != 0 {
		t.Errorf("Envio(900,00) = %d", got)
	}
}

// Este «test» ejecuta Etiqueta con las dos ramas y no comprueba nada. También cuenta para
// la cobertura.
func TestEtiqueta(t *testing.T) {
	Etiqueta(10000)
	Etiqueta(90000)
}
