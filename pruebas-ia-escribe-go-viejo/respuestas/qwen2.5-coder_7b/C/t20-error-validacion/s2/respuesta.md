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
	return "Campo " + e.Campo + " invalido"
}

func CampoInvalido(err error) (string, bool) {
	var ev ErrValidacion
	return ev.Campo, errors.As(err, &ev)
}
```