package pedidos

import "testing"

func TestProcesarAvisaDelPanico(t *testing.T) {
	hechos := 0
	var descuentos map[string]int // nadie la inicializó
	err := Procesar([]func(){
		func() { hechos++ },                         // cobrar
		func() { descuentos["ana"] = 10; hechos++ }, // pánico: mapa nil
		func() { hechos++ },                         // enviar
	})
	if err == nil {
		t.Fatalf("Procesar -> err = nil, pero solo se hicieron %d de 3 pasos: cobrado y sin enviar", hechos)
	}
	t.Logf("Procesar -> %v (pasos hechos: %d de 3)", err, hechos)
}
