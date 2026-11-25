//go:build malo

package pedidos

import "fmt"

// Exportar manda cada pedido al almacén por una conexión del pool.
func Exportar(p *Pool, ids []int) error {
	for _, id := range ids {
		c, err := p.Tomar()
		if err != nil {
			return fmt.Errorf("pedido %d: %w", id, err)
		}
		defer c.Soltar()
		c.Escribir(id)
	}
	return nil
}
