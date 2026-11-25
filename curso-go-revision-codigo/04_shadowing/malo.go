//go:build malo

package pedidos

import "strconv"

// Total suma las cantidades de las líneas del pedido.
func Total(cantidades []string) (int, error) {
	var err error
	total := 0
	for _, c := range cantidades {
		n, err := strconv.Atoi(c)
		if err != nil {
			break
		}
		total += n
	}
	return total, err
}
