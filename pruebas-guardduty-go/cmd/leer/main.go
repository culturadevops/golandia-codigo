// leer: pasa los eventos de testdata/eventos.json por la misma lógica que el handler, y enseña
// qué sale de cada uno. Los eventos están construidos a mano con los campos que documenta AWS:
// no son de una cuenta real.
package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"os"

	"github.com/aws/aws-lambda-go/events"

	"github.com/culturadevops/pruebas-guardduty-go/alerta"
)

func main() {
	b, err := os.ReadFile("testdata/eventos.json")
	if err != nil {
		log.Fatal(err)
	}
	var evs []events.CloudWatchEvent
	if err := json.Unmarshal(b, &evs); err != nil {
		log.Fatal(err)
	}
	for _, ev := range evs {
		d, err := alerta.Decidir(ev)
		if err != nil {
			log.Fatal(err)
		}
		t := d.Tipo
		fmt.Println(d.Finding.Type)
		fmt.Printf("  propósito=%s recurso=%s familia=%s", t.Proposito, t.Recurso, t.Familia)
		if t.Mecanismo != "" {
			fmt.Printf(" mecanismo=%s", t.Mecanismo)
		}
		if t.Artefacto != "" {
			fmt.Printf(" artefacto=%s", t.Artefacto)
		}
		fmt.Println()
		fmt.Printf("  severidad=%.1f → %s · detectó: %s · veces=%d\n",
			d.Finding.Severity, d.Tramo, t.Origen(), d.Finding.Service.Count)
		fmt.Printf("  acción: %s\n\n", d.Accion)
	}
	m, _ := alerta.Mensaje(mustDecidir(evs[1]))
	fmt.Println("mensaje para la notificación (la descripción trae un salto de línea):")
	fmt.Println(string(m))
	fmt.Println("¿JSON válido?", json.Valid(m))

	// json.Marshal escapa <, > y & para que el JSON sea seguro dentro de HTML.
	// Con un Encoder y SetEscapeHTML(false), no.
	var buf bytes.Buffer
	enc := json.NewEncoder(&buf)
	enc.SetEscapeHTML(false)
	if err := enc.Encode(map[string]string{"cuenta": "<CUENTA>"}); err != nil {
		log.Fatal(err)
	}
	fmt.Print("con SetEscapeHTML(false): ", buf.String())
}

func mustDecidir(ev events.CloudWatchEvent) alerta.Decision {
	d, err := alerta.Decidir(ev)
	if err != nil {
		log.Fatal(err)
	}
	return d
}
