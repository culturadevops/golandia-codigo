package tarea

import (
	"fmt"
	"strings"
)

func Describir(values ...any) []string {
	var result []string
	for _, value := range values {
		result = append(result, fmt.Sprintf("%T", value))
	}
	return result
}
