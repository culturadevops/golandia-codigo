//go:build malo

package pedidos

import "encoding/json"

type Pedido struct {
	Cliente  string `json:"cliente"`
	Cantidad int    `json:"cantidad"`
}

// Decodificar lee el cuerpo de la petición POST /pedidos.
func Decodificar(cuerpo []byte) (Pedido, error) {
	var p Pedido
	json.Unmarshal(cuerpo, &p)
	return p, nil
}
