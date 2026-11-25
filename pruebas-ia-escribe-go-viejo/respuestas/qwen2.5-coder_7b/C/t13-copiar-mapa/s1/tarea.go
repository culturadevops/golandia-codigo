package tarea

import (
	"strings"
	"sync"
)

func Mezclar(dst, src map[string]int) {
	for k, v := range src {
		dst[k] = v
	}
}
