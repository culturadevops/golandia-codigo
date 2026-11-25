// Package supresion crea reglas de supresión de GuardDuty, y se niega a crear las peligrosas.
//
// En GuardDuty, una regla de supresión es un filtro con acción ARCHIVE: archiva sola los findings
// nuevos que cumplen el criterio. La regla de oro de la documentación: reactivas y estrechas,
// siempre tipo + un recurso concreto. Una regla demasiado amplia deja ciega a la detección de
// secuencias de ataque, porque los findings archivados no cuentan como señales. Y no avisa.
package supresion

import (
	"errors"
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/guardduty"
	"github.com/aws/aws-sdk-go-v2/service/guardduty/types"
)

// Regla es una supresión: un tipo de finding y los criterios que acotan el recurso.
type Regla struct {
	Nombre      string
	Descripcion string
	Tipo        string            // p. ej. Recon:EC2/Portscan
	Recurso     map[string]string // atributo del finding → valor exacto
}

var (
	ErrSinTipo    = errors.New("una supresión sin tipo archiva findings de cualquier tipo")
	ErrSinRecurso = errors.New("una supresión solo por tipo apaga ese tipo en TODA la cuenta")
	ErrComodin    = errors.New("un tipo con * no es un tipo concreto")
)

// Validar aplica la regla de oro antes de gastar una llamada a la API.
func (r Regla) Validar() error {
	switch {
	case r.Tipo == "":
		return ErrSinTipo
	case strings.Contains(r.Tipo, "*"):
		return ErrComodin
	case len(r.Recurso) == 0:
		return ErrSinRecurso
	}
	return nil
}

// Entrada construye el CreateFilterInput con acción ARCHIVE.
func (r Regla) Entrada(detector string, rango int32) (*guardduty.CreateFilterInput, error) {
	if err := r.Validar(); err != nil {
		return nil, fmt.Errorf("%s: %w", r.Nombre, err)
	}
	crit := map[string]types.Condition{"type": {Equals: []string{r.Tipo}}}
	for k, v := range r.Recurso {
		crit[k] = types.Condition{Equals: []string{v}}
	}
	return &guardduty.CreateFilterInput{
		DetectorId:      aws.String(detector),
		Name:            aws.String(r.Nombre),
		Description:     aws.String(r.Descripcion),
		Action:          types.FilterActionArchive,
		Rank:            aws.Int32(rango),
		FindingCriteria: &types.FindingCriteria{Criterion: crit},
	}, nil
}
