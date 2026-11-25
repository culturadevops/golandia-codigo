//go:build !malo

package pedidos

import (
	"fmt"
	"strconv"
)

// Total suma las cantidades de las líneas del pedido.
func Total(cantidades []string) (int, error) {
	total := 0
	for i, c := range cantidades {
		n, err := strconv.Atoi(c)
		if err != nil {
			return 0, fmt.Errorf("línea %d: %w", i+1, err)
		}
		total += n
	}
	return total, nil
}
