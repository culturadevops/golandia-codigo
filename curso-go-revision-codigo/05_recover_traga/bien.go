//go:build !malo

package pedidos

import "fmt"

// Procesar corre los pasos del pedido. Si uno entra en pánico, el servicio no se cae,
// pero quien llama se entera: el pánico se convierte en un error.
func Procesar(pasos []func()) (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("pánico procesando el pedido: %v", r)
		}
	}()
	for _, paso := range pasos {
		paso()
	}
	return nil
}
