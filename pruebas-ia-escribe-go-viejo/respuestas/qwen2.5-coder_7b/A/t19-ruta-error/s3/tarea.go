package tarea

import (
	"errors"
	"io/fs"
	"strings"
)

func RutaDelError(err error) (string, bool) {
	var pathError *fs.PathError
	if errors.As(err, &pathError) {
		return pathError.Path, true
	}
	return "", false
}
