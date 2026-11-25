//go:build !malo

package pedidos

import (
	"io"
	"net/http"
)

// Disponible pregunta al almacén si un producto tiene stock (200 = sí).
func Disponible(c *http.Client, url string) (bool, error) {
	resp, err := c.Get(url)
	if err != nil {
		return false, err
	}
	defer resp.Body.Close()
	_, _ = io.Copy(io.Discard, resp.Body) // leerlo entero deja reutilizar la conexión
	return resp.StatusCode == http.StatusOK, nil
}
