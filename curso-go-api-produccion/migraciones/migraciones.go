// Package migraciones mete los .sql dentro del binario: la imagen no necesita la carpeta.
package migraciones

import (
	"embed"
	"io/fs"
	"strconv"
	"strings"
)

//go:embed *.sql
var FS embed.FS

// Ultima es la versión del esquema que espera ESTE binario: la del .sql más alto.
// La readiness la compara con la de la base de datos.
func Ultima() int64 {
	var max int64
	nombres, _ := fs.Glob(FS, "*.sql")
	for _, n := range nombres {
		v, err := strconv.ParseInt(strings.SplitN(n, "_", 2)[0], 10, 64)
		if err == nil && v > max {
			max = v
		}
	}
	return max
}
