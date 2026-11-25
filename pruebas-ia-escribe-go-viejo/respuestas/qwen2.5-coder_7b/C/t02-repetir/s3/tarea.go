package tarea

import (
	"strings"
)

func Repetir(s string, n int) []string {
	result := make([]string, n)
	for i := range n {
		result[i] = s
	}
	return result
}
