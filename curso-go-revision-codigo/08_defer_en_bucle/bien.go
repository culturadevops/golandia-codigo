//go:build !malo

package pedidos

import "fmt"

// Exportar manda cada pedido al almacén por una conexión del pool.
func Exportar(p *Pool, ids []int) error {
	for _, id := range ids {
		if err := exportarUno(p, id); err != nil {
			return fmt.Errorf("pedido %d: %w", id, err)
		}
	}
	return nil
}

// exportarUno: el defer corre al salir de ESTA función, en cada vuelta.
func exportarUno(p *Pool, id int) error {
	c, err := p.Tomar()
	if err != nil {
		return err
	}
	defer c.Soltar()
	c.Escribir(id)
	return nil
}
