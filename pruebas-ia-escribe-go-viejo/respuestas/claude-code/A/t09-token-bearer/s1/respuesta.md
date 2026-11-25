```go
package tarea

import "strings"

// TokenBearer devuelve el token si la cabecera empieza por "Bearer ".
func TokenBearer(cabecera string) (string, bool) {
	const prefijo = "Bearer "
	if !strings.HasPrefix(cabecera, prefijo) {
		return "", false
	}
	return strings.TrimPrefix(cabecera, prefijo), true
}
```