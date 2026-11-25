//go:build malo

package pedidos

import "strconv"

// Cantidad lee la cantidad del formulario; el almacén la guarda en un int16.
func Cantidad(s string) (int16, error) {
	n, err := strconv.Atoi(s)
	if err != nil {
		return 0, err
	}
	return int16(n), nil
}
