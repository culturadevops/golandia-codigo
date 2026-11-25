```go
package tarea

import "maps"

// Mezclar copia todas las entradas de src en dst, sobrescribiendo las claves existentes.
func Mezclar(dst, src map[string]int) {
	maps.Copy(dst, src)
}
```