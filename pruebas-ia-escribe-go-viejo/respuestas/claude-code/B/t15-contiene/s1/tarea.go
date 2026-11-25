package tarea

import "slices"

// EsAdmin devuelve true si roles contiene el texto "admin".
func EsAdmin(roles []string) bool {
	return slices.Contains(roles, "admin")
}
