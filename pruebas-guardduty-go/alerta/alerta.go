// Package alerta decide qué hacer con un evento de GuardDuty que llega por EventBridge.
//
// Es la lógica de un handler de Lambda, separada del runtime para poder probarla con go test
// sin una cuenta de AWS. El handler de verdad está en cmd/lambda.
package alerta

import (
	"encoding/json"
	"fmt"

	"github.com/aws/aws-lambda-go/events"

	"github.com/culturadevops/pruebas-guardduty-go/hallazgo"
)

// Finding son los campos del detalle que usamos. El finding completo trae muchos más.
type Finding struct {
	ID          string  `json:"id"`
	Type        string  `json:"type"`
	Severity    float64 `json:"severity"` // un número decimal, no una palabra
	Title       string  `json:"title"`
	Description string  `json:"description"`
	AccountID   string  `json:"accountId"`
	Region      string  `json:"region"`
	UpdatedAt   string  `json:"updatedAt"`
	Service     struct {
		ResourceRole   string `json:"resourceRole"`
		Count          int    `json:"count"`
		AdditionalInfo struct {
			Sample bool `json:"sample"`
		} `json:"additionalInfo"`
	} `json:"service"`
}

// Accion es lo que el handler decide.
type Accion string

const (
	Ignorar  Accion = "ignorar"   // no es un finding de GuardDuty
	Probar   Accion = "prueba"    // un finding de muestra: sirve para probar la tubería
	Registro Accion = "registro"  // Low: tomar nota
	Cola     Accion = "cola"      // Medium: investigar pronto
	Avisar   Accion = "avisar-ya" // High y Critical
)

// Decision es lo que sale del handler.
type Decision struct {
	Accion  Accion
	Tramo   hallazgo.Tramo
	Tipo    hallazgo.Tipo
	Finding Finding
}

// Decidir lee un evento de EventBridge y decide.
func Decidir(ev events.CloudWatchEvent) (Decision, error) {
	if ev.Source != "aws.guardduty" || ev.DetailType != "GuardDuty Finding" {
		return Decision{Accion: Ignorar}, nil
	}
	var f Finding
	if err := json.Unmarshal(ev.Detail, &f); err != nil {
		return Decision{}, fmt.Errorf("detalle ilegible: %w", err)
	}
	tipo, err := hallazgo.ParseTipo(f.Type)
	if err != nil {
		return Decision{}, err
	}
	tramo, err := hallazgo.TramoDe(f.Severity)
	if err != nil {
		return Decision{}, err
	}
	d := Decision{Tramo: tramo, Tipo: tipo, Finding: f}
	switch {
	case f.Service.AdditionalInfo.Sample:
		d.Accion = Probar
	case tramo == hallazgo.Critical || tramo == hallazgo.High:
		d.Accion = Avisar
	case tramo == hallazgo.Medium:
		d.Accion = Cola
	default:
		d.Accion = Registro
	}
	return d, nil
}

// Mensaje arma el cuerpo JSON de la notificación.
//
// La documentación avisa: una plantilla de SNS en varias líneas, con una descripción que trae
// saltos de línea, da un JSON inválido y la notificación no sale. json.Marshal escapa el salto
// como \n, así que aquí no puede pasar.
func Mensaje(d Decision) ([]byte, error) {
	return json.Marshal(map[string]any{
		"accion":      d.Accion,
		"tramo":       d.Tramo,
		"tipo":        d.Finding.Type,
		"recurso":     d.Tipo.Recurso,
		"proposito":   d.Tipo.Proposito,
		"cuenta":      d.Finding.AccountID,
		"region":      d.Finding.Region,
		"veces":       d.Finding.Service.Count,
		"descripcion": d.Finding.Description,
	})
}
