```go
package tarea

import (
	"strings"
)

func ParClaveValor(s string) (clave, valor string, ok bool) {
	parts := strings.Cut(s, "=")
	if parts[1] != "" {
		return parts[0], parts[1], true
	}
	return "", "", false
}
```