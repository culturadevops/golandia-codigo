//go:build !malo

package pedidos

import (
	"encoding/json"
	"fmt"
)

type Pedido struct {
	Cliente  string `json:"cliente"`
	Cantidad int    `json:"cantidad"`
}

// Decodificar lee el cuerpo de la petición POST /pedidos.
func Decodificar(cuerpo []byte) (Pedido, error) {
	var p Pedido
	if err := json.Unmarshal(cuerpo, &p); err != nil {
		return Pedido{}, fmt.Errorf("pedido ilegible: %w", err)
	}
	return p, nil
}
