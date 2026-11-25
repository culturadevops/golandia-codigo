package tarea

import (
	"fmt"
)

func Describir(valores ...interface{}) []string {
	descripciones := make([]string, len(valores))
	for i, valor := range valores {
		descripciones[i] = fmt.Sprintf("%T", valor)
	}
	return descripciones
}
