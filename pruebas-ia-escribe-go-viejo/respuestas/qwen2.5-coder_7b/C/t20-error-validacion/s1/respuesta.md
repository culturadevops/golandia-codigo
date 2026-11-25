```go
package tarea

import (
	"errors"
	"strings"
)

type ErrValidacion struct {
	Campo string
}

func (e ErrValidacion) Error() string {
	return "error de validación en campo: " + e.Campo
}

func CampoInvalido(err error) (string, bool) {
	var ve ErrValidacion
	if errors.As(err, &ve) {
		return ve.Campo, true
	}
	return "", false
}
```