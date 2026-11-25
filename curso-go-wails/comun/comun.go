// Package comun: lo que comparten las lecciones. Reportar es el «chivato»: el frontend le cuenta a Go
// lo que ha recibido por el puente, y Go lo escribe en la salida estándar. Así lo que ve JavaScript
// queda en SALIDA.md, literal, y no hace falta leerlo de una captura.
package comun

import "fmt"

type Chivato struct{}

func (Chivato) Reportar(linea string) { fmt.Println(linea) }
