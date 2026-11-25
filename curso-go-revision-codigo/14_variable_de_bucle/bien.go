//go:build !malo

// El mismo código, con las reglas del go.mod (go 1.25): desde Go 1.22 cada vuelta
// del bucle tiene su propia variable. El arreglo es subir el go.mod; en un repo que
// no puede subirlo, se copia la variable a mano: id := id.

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
