package tarea

import (
	"errors"
	"strings"
)

type ErrValidacion struct {
	Campo string
}

func (e *ErrValidacion) Error() string {
	return "Error de validación en campo " + e.Campo
}

func CampoInvalido(err error) (string, bool) {
	if errors.As(err, &ErrValidacion{Campo: ""}) {
		return "", false
	}
	errStr := err.Error()
	for _, part := range strings.Split(errStr, " ") {
		if strings.HasPrefix(part, "campo") {
			campo := strings.TrimPrefix(part, "campo ")
			return campo, true
		}
	}
	return "", false
}
