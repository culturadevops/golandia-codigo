```go
package tarea

import "strings"

// ParClaveValor separa s por el primer signo igual en clave y valor.
// Si s no contiene ningún signo igual, ok es false.
func ParClaveValor(s string) (clave, valor string, ok bool) {
	return strings.Cut(s, "=")
}
```