package tarea

import (
	"strings"
)

func LineaTabla(celdas []string) string {
	var builder strings.Builder
	builder.WriteString("|")
	for i, celda := range celdas {
		builder.WriteString(" ")
		builder.WriteString(celda)
		builder.WriteString(" |")
		if i < len(celdas)-1 {
			builder.WriteString(" ")
		}
	}
	return builder.String()
}
