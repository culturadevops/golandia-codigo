//go:build !malo

package pedidos

import (
	"fmt"
	"strconv"
)

// Cantidad lee la cantidad del formulario; el almacén la guarda en un int16.
// ParseInt con bitSize 16 devuelve error si no cabe, en vez de dar la vuelta.
func Cantidad(s string) (int16, error) {
	n, err := strconv.ParseInt(s, 10, 16)
	if err != nil {
		return 0, fmt.Errorf("cantidad %q: %w", s, err)
	}
	if n <= 0 {
		return 0, fmt.Errorf("cantidad %q: tiene que ser mayor que cero", s)
	}
	return int16(n), nil
}
