//go:build malo && go1.21

// La línea de arriba dice «go1.21»: este archivo se compila con las reglas de Go 1.21,
// igual que todo un repo cuyo go.mod diga «go 1.21». Antes de Go 1.22, el bucle tenía
// UNA sola variable id para todas las vueltas.

package pedidos

import "fmt"

// Avisos prepara un aviso por pedido; se mandan después, todos juntos.
func Avisos(ids []int) []func() string {
	var avisos []func() string
	for _, id := range ids {
		avisos = append(avisos, func() string { return fmt.Sprintf("pedido %d listo", id) })
	}
	return avisos
}
