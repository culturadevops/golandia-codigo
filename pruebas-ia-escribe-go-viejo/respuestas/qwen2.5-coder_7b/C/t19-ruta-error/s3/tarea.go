package tarea

import (
	"errors"
	"fmt"
	"path/filepath"
	"strings"
)

func RutaDelError(err error) (string, bool) {
	var pathErr *fs.PathError
	if errors.As(err, &pathErr) {
		return pathErr.Path, true
	}
	return "", false
}
