//go:build !malo

package pedidos

import "time"

// Recoger junta las confirmaciones de los almacenes durante un plazo total.
// El temporizador se crea UNA vez, fuera del bucle.
func Recoger(ch <-chan string, plazo time.Duration) []string {
	var r []string
	fin := time.NewTimer(plazo)
	defer fin.Stop()
	for {
		select {
		case m := <-ch:
			r = append(r, m)
		case <-fin.C:
			return r
		}
	}
}
