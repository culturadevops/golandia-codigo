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
	return "Campo " + e.Campo + " invalido"
}

func CampoInvalido(err error) (string, bool) {
	if err == nil {
		return "", false
	}

	errStr := err.Error()
	if strings.HasPrefix(errStr, "Campo ") && strings.HasSuffix(errStr, " invalido") {
		campo := strings.TrimPrefix(strings.TrimSuffix(errStr, " invalido"), "Campo ")
		return campo, true
	}

	return "", false
}
```