// Paquete de una biblioteca interna que cambió de API.
package viejo

import "strings"

// Normalizar es la función nueva.
func Normalizar(s string) string { return strings.ToLower(strings.TrimSpace(s)) }

// Limpiar se queda por compatibilidad.
//
// Deprecated: usa Normalizar.
//
//go:fix inline
func Limpiar(s string) string { return Normalizar(s) }
