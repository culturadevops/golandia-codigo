//go:build malo

package pedidos

type Pedido struct {
	Cliente  string
	Cantidad int
}

type ErrValidacion struct{ Campo string }

func (e *ErrValidacion) Error() string { return "campo inválido: " + e.Campo }

// Validar devuelve nil si el pedido está bien.
func Validar(p Pedido) error {
	var e *ErrValidacion
	if p.Cantidad <= 0 {
		e = &ErrValidacion{Campo: "cantidad"}
	}
	return e
}
