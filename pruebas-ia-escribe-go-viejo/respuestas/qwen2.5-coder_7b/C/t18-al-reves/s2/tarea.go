package tarea

import (
	"strings"
)

func AlReves(xs []string) []string {
	n := len(xs)
	result := make([]string, n)
	for i := range xs {
		result[n-i-1] = xs[i]
	}
	return result
}
