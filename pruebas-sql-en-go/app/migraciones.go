// Package tienda guarda las migraciones dentro del binario, para los tests.
package tienda

import "embed"

//go:embed migrations/*.sql
var Migraciones embed.FS
