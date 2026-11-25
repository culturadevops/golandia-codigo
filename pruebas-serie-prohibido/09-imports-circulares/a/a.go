// Video 207 — los imports circulares son ERROR DE COMPILACION en Go.
package a

import "ciclo/b"

type Pedido struct{ ID string }

func Hacer() { b.Cobrar() }
