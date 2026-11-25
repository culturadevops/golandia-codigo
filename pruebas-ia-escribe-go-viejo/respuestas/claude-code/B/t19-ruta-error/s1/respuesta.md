```go
package tarea

import (
	"errors"
	"io/fs"
)

// RutaDelError devuelve el campo Path del primer *fs.PathError encontrado
// en la cadena de errores de err, junto con true. Si no hay ninguno,
// devuelve "" y false.
func RutaDelError(err error) (string, bool) {
	if pe, ok := errors.AsType[*fs.PathError](err); ok && pe != nil {
		return pe.Path, true
	}
	return "", false
}
```