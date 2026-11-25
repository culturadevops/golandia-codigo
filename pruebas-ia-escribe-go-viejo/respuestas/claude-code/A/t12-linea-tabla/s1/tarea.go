package tarea

import "strings"

// LineaTabla devuelve las celdas unidas en formato de fila de tabla: "| a | b | c |".
func LineaTabla(celdas []string) string {
	var sb strings.Builder
	sb.WriteString("|")
	for _, celda := range celdas {
		sb.WriteString(" ")
		sb.WriteString(celda)
		sb.WriteString(" |")
	}
	return sb.String()
}
