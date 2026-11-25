//go:build malo

package pedidos

import "time"

// Recoger junta las confirmaciones de los almacenes durante un plazo total.
func Recoger(ch <-chan string, plazo time.Duration) []string {
	var r []string
	for {
		select {
		case m := <-ch:
			r = append(r, m)
		case <-time.After(plazo):
			return r
		}
	}
}
