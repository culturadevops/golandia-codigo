package alerta

import (
	"encoding/json"
	"os"
	"strings"
	"testing"

	"github.com/aws/aws-lambda-go/events"
)

func cargar(t *testing.T) []events.CloudWatchEvent {
	t.Helper()
	b, err := os.ReadFile("../testdata/eventos.json")
	if err != nil {
		t.Fatal(err)
	}
	var evs []events.CloudWatchEvent
	if err := json.Unmarshal(b, &evs); err != nil {
		t.Fatal(err)
	}
	return evs
}

func TestDecidir(t *testing.T) {
	quiere := map[string]Accion{
		"hallazgo-portprobe":    Registro, // severidad 2: Low
		"hallazgo-exfiltracion": Avisar,   // 8: High
		"hallazgo-muestra":      Probar,   // 8, pero es [SAMPLE]
		"hallazgo-secuencia":    Avisar,   // 9.2: Critical
	}
	for _, ev := range cargar(t) {
		d, err := Decidir(ev)
		if err != nil {
			t.Fatal(err)
		}
		if d.Accion != quiere[d.Finding.ID] {
			t.Errorf("%s (%s %.1f): %s, quiere %s", d.Finding.ID, d.Finding.Type,
				d.Finding.Severity, d.Accion, quiere[d.Finding.ID])
		}
	}
}

func TestIgnoraOtrosEventos(t *testing.T) {
	ev := events.CloudWatchEvent{Source: "aws.ec2", DetailType: "EC2 Instance State-change Notification"}
	d, err := Decidir(ev)
	if err != nil || d.Accion != Ignorar {
		t.Fatalf("got %v, %v", d.Accion, err)
	}
}

// La descripción del evento 2 trae un salto de línea. El mensaje tiene que seguir siendo JSON.
func TestMensajeConSaltoDeLinea(t *testing.T) {
	d, err := Decidir(cargar(t)[1])
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(d.Finding.Description, "\n") {
		t.Fatal("el caso de prueba debería traer un salto de línea")
	}
	m, err := Mensaje(d)
	if err != nil {
		t.Fatal(err)
	}
	if !json.Valid(m) {
		t.Fatalf("JSON inválido: %s", m)
	}
}
