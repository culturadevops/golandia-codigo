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
	if err == nil {
		return "", false
	}

	if errStr, ok := err.(*ErrValidacion); ok {
		return errStr.Campo, true
	}

	if strings.Contains(err.Error(), "Error de validación en campo: ") {
		campo := strings.Split(err.Error(), ": ")[1]
		return campo, true
	}

	return "", false
}
```