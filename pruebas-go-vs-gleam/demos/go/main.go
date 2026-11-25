// Los estados de un pedido, y un contador de 64 bits que llega al tope.
package main

import (
	"fmt"
	"math"
)

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
	var tope uint64 = math.MaxUint64
	fmt.Println("2^64 + 1 =", tope+2)
}
