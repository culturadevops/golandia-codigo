```go
package tarea

// Mezclar copia todas las entradas de src en dst, sobrescribiendo las claves
// que ya existan. dst no debe ser nil si src tiene entradas.
func Mezclar(dst, src map[string]int) {
	for k, v := range src {
		dst[k] = v
	}
}
```