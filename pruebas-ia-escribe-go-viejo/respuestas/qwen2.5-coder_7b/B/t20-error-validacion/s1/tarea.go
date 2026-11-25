package tarea

import (
	"errors"
	"strings"
)

type ErrValidacion struct {
	Campo string
}

func (e *ErrValidacion) Error() string {
	return "campo " + e.Campo + " invalido"
}

func CampoInvalido(err error) (string, bool) {
	if err == nil {
		return "", false
	}

	errStr := err.Error()
	if strings.HasPrefix(errStr, "campo ") && strings.HasSuffix(errStr, " invalido") {
		campo := strings.TrimPrefix(errStr, "campo ")
		campo = strings.TrimSuffix(campo, " invalido")
		return campo, true
	}

	return "", false
}
