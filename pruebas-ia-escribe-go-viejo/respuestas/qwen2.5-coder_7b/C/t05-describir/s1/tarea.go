package tarea

import (
	"fmt"
	"strings"
)

func Describir(valores ...any) []string {
	var tipos []string
	for _, valor := range valores {
		tipos = append(tipos, fmt.Sprintf("%T", valor))
	}
	return tipos
}
