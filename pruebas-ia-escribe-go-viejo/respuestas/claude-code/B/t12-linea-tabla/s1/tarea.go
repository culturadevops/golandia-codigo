package tarea

import "strings"

// LineaTabla devuelve las celdas unidas en formato de fila de tabla: "| a | b | c |".
func LineaTabla(celdas []string) string {
	var b strings.Builder
	b.WriteString("|")
	for _, celda := range celdas {
		b.WriteString(" ")
		b.WriteString(celda)
		b.WriteString(" |")
	}
	return b.String()
}
