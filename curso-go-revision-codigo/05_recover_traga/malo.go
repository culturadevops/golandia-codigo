//go:build malo

package pedidos

import "log"

// Procesar corre los pasos del pedido. Si uno entra en pánico, el servicio no se cae.
func Procesar(pasos []func()) error {
	defer func() {
		if r := recover(); r != nil {
			log.Printf("recuperado: %v", r)
		}
	}()
	for _, paso := range pasos {
		paso()
	}
	return nil
}
