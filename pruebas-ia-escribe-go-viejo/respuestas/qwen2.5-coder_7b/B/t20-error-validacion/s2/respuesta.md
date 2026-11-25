```go
package tarea

import (
	"errors"
	"strings"
)

type ErrValidacion struct {
	Campo string
}

func (e *ErrValidacion) Error() string {
	return "Error de validación en campo: " + e.Campo
}

func CampoInvalido(err error) (string, bool) {
	var ev *ErrValidacion
	if errors.As(err, &ev) {
		return ev.Campo, true
	}
	return "", false
}
```