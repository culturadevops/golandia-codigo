```go
package tarea

// EsAdmin devuelve true si el slice de roles contiene el texto "admin".
func EsAdmin(roles []string) bool {
	for _, rol := range roles {
		if rol == "admin" {
			return true
		}
	}
	return false
}
```