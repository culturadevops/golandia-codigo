//go:build malo

package pedidos

import "net/http"

// Disponible pregunta al almacén si un producto tiene stock (200 = sí).
func Disponible(c *http.Client, url string) (bool, error) {
	resp, err := c.Get(url)
	if err != nil {
		return false, err
	}
	return resp.StatusCode == http.StatusOK, nil
}
