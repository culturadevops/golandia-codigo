package tarea

import (
	"fmt"
)

func Describir(valores ...interface{}) []string {
	resultado := make([]string, len(valores))
	for i, valor := range valores {
		resultado[i] = fmt.Sprintf("%T", valor)
	}
	return resultado
}
