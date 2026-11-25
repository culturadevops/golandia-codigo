// Package malo es la línea de la lección 1.2, aparte, para que go vet la juzgue sola.
package malo

import "log/slog"

func Rechazar() {
	slog.Info("pago rechazado", "pedido", 42, "tarjeta caducada")
}
