// Los estados de un pedido. Alguien agregó Devuelto y nadie tocó la función que los describe.
package main

import "fmt"

type Estado int

const (
	Pendiente Estado = iota
	Pagado
	Enviado
	Devuelto
)

func describir(e Estado) string {
	switch e {
	case Pendiente:
		return "esperando el pago"
	case Pagado:
		return "listo para enviar"
	case Enviado:
		return "en camino"
	}
	return ""
}

func main() {
	fmt.Printf("Devuelto: %q\n", describir(Devuelto))
}
