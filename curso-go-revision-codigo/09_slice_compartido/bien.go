//go:build !malo

package pedidos

import "slices"

// ConEnvio devuelve las líneas del carrito más la línea del envío elegido.
// slices.Clip deja la capacidad igual a la longitud: el append tiene que copiar.
func ConEnvio(lineas []string, envio string) []string {
	return append(slices.Clip(lineas), envio)
}
