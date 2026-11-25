//go:build malo

package pedidos

import "encoding/json"

// Recibo es lo que devuelve GET /pedidos/{id}/recibo.
type Recibo struct {
	id    int     `json:"id"`
	total float64 `json:"total"`
}

func NuevoRecibo(id int, total float64) Recibo { return Recibo{id: id, total: total} }

func (r Recibo) JSON() ([]byte, error) { return json.Marshal(r) }
