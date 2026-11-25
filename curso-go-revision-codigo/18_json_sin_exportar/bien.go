//go:build !malo

package pedidos

import "encoding/json"

// Recibo es lo que devuelve GET /pedidos/{id}/recibo.
// encoding/json solo ve los campos exportados (con mayúscula); la etiqueta pone el nombre.
type Recibo struct {
	ID    int     `json:"id"`
	Total float64 `json:"total"`
}

func NuevoRecibo(id int, total float64) Recibo { return Recibo{ID: id, Total: total} }

func (r Recibo) JSON() ([]byte, error) { return json.Marshal(r) }
