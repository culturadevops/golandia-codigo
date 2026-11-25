```go
package tarea

import "strings"

// TokenBearer devuelve el texto que sigue a "Bearer " y true si la cabecera
// empieza por ese prefijo; en caso contrario devuelve "" y false.
func TokenBearer(cabecera string) (string, bool) {
	return strings.CutPrefix(cabecera, "Bearer ")
}
```