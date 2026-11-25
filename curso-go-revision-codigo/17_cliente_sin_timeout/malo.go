//go:build malo

package pedidos

import (
	"io"
	"net/http"
)

var cliente = &http.Client{}

// Cotizar le pide el costo del envío a la paquetería.
func Cotizar(url string) (string, error) {
	resp, err := cliente.Get(url)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	b, err := io.ReadAll(resp.Body)
	return string(b), err
}
