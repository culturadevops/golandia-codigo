// Package hallazgo lee lo que GuardDuty ya te dice en el nombre y en la severidad de un finding.
//
// El formato del tipo, según la documentación de GuardDuty:
//
//	ThreatPurpose:ResourceTypeAffected/ThreatFamilyName.DetectionMechanism!Artifact
//
// El mecanismo (después del primer punto) y el artefacto (después de «!») son opcionales.
package hallazgo

import (
	"fmt"
	"strings"
)

// Tipo es el nombre de un finding, partido en sus piezas.
type Tipo struct {
	Proposito string // Recon, Backdoor, UnauthorizedAccess...
	Recurso   string // EC2, IAMUser, S3, Kubernetes, Runtime...
	Familia   string // PortProbeUnprotectedPort, C&CActivity...
	Mecanismo string // B, Tcp, Custom, Reputation, OutsideAWS... (opcional)
	Artefacto string // DNS... (opcional)
}

// ParseTipo parte el nombre de un tipo de finding. Falla si no tiene «:» o «/».
func ParseTipo(s string) (Tipo, error) {
	var t Tipo
	proposito, resto, ok := strings.Cut(s, ":")
	if !ok || proposito == "" {
		return t, fmt.Errorf("tipo %q: falta «Proposito:»", s)
	}
	recurso, resto, ok := strings.Cut(resto, "/")
	if !ok || recurso == "" || resto == "" {
		return t, fmt.Errorf("tipo %q: falta «Recurso/Familia»", s)
	}
	t.Proposito, t.Recurso = proposito, recurso
	if i := strings.LastIndex(resto, "!"); i >= 0 {
		resto, t.Artefacto = resto[:i], resto[i+1:]
	}
	t.Familia, t.Mecanismo, _ = strings.Cut(resto, ".")
	return t, nil
}

// Origen explica los dos sufijos que cambian quién detectó el finding.
func (t Tipo) Origen() string {
	switch t.Mecanismo {
	case "Custom":
		return "tu propia lista de amenazas"
	case "Reputation":
		return "el modelo de reputación de dominios"
	}
	return "la inteligencia de AWS"
}

// Tramo es la severidad en palabras.
type Tramo string

const (
	Low      Tramo = "Low"
	Medium   Tramo = "Medium"
	High     Tramo = "High"
	Critical Tramo = "Critical"
)

// TramoDe convierte la severidad numérica (1.0 a 10.0) en su tramo:
// Low 1.0-3.9 · Medium 4.0-6.9 · High 7.0-8.9 · Critical 9.0-10.0.
func TramoDe(sev float64) (Tramo, error) {
	switch {
	case sev < 1 || sev > 10:
		return "", fmt.Errorf("severidad %.1f fuera de 1.0-10.0", sev)
	case sev >= 9:
		return Critical, nil
	case sev >= 7:
		return High, nil
	case sev >= 4:
		return Medium, nil
	}
	return Low, nil
}
