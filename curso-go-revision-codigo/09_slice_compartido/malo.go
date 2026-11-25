//go:build malo

package pedidos

// ConEnvio devuelve las líneas del carrito más la línea del envío elegido.
func ConEnvio(lineas []string, envio string) []string {
	return append(lineas, envio)
}
