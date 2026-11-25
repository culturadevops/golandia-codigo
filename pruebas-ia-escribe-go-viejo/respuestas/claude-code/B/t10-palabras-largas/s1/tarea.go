package tarea

import (
	"strings"
	"unicode"
)

// PalabrasLargas cuenta cuántas palabras del texto, separadas por espacios
// en blanco, tienen más de 3 letras.
func PalabrasLargas(texto string) int {
	cuenta := 0
	for _, palabra := range strings.Fields(texto) {
		letras := 0
		for _, r := range palabra {
			if unicode.IsLetter(r) {
				letras++
			}
		}
		if letras > 3 {
			cuenta++
		}
	}
	return cuenta
}
