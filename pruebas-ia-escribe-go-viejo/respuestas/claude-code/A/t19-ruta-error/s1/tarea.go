package tarea

import (
	"errors"
	"io/fs"
)

// RutaDelError busca un *fs.PathError en la cadena de errores de err.
// Si lo encuentra, devuelve su campo Path y true; si no, "" y false.
func RutaDelError(err error) (string, bool) {
	var pathErr *fs.PathError
	if errors.As(err, &pathErr) && pathErr != nil {
		return pathErr.Path, true
	}
	return "", false
}
