// sintemporal: el MISMO flujo (sin los fallos del cobro) como un programa normal de Go.
// Si el proceso muere, al volver a lanzarlo no sabe por dónde iba: empieza de cero.
package main

import (
	"fmt"
	"os"
	"time"

	"temporaldemo/pedido"
)

func paso(f string, a ...any) {
	fmt.Printf("%s [sin-temporal] %s\n", time.Now().UTC().Format("15:04:05.000"), fmt.Sprintf(f, a...))
}

func main() {
	id := os.Args[1]
	q := "sin-temporal"
	paso("reservo el stock de %s", id)
	time.Sleep(time.Second)
	pedido.Efecto(q, "reservar "+id)
	paso("cobro %s", id)
	pedido.Efecto(q, "cobrar "+id)
	for p := 1; p <= pedido.Paginas; p++ {
		time.Sleep(time.Second)
		pedido.Efecto(q, fmt.Sprintf("factura %s página %d", id, p))
		paso("página %d de %d", p, pedido.Paginas)
	}
	paso("envío %s", id)
	pedido.Efecto(q, "enviar "+id)
}
