```go
package tarea

import (
	"strings"
)

func PalabrasLargas(texto string) int {
	palabras := strings.Fields(texto)
	cuenta := 0
	for _, palabra := range palabras {
		if len(palabra) > 3 {
			cuenta++
		}
	}
	return cuenta
}
```