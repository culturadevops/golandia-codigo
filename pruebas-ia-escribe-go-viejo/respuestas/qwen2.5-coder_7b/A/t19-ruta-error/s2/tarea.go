package tarea

import (
	"errors"
	"io/fs"
)

func RutaDelError(err error) (string, bool) {
	var pathErr *fs.PathError
	if errors.As(err, &pathErr) {
		return pathErr.Path, true
	}
	return "", false
}
