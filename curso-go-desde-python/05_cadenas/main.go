// Leccion 5 — Cadenas: len() cuenta BYTES, no letras.
package main

import (
	"fmt"
	"unicode/utf8"
)

func main() {
	s := "configuración"
	fmt.Println("== 5.1 · La misma palabra que en el curso de Python ==")
	fmt.Printf("  len(%q)                  -> %d   <- BYTES\n", s, len(s))
	fmt.Printf("  utf8.RuneCountInString   -> %d   <- letras\n", utf8.RuneCountInString(s))
	fmt.Printf("  En Python len() devuelve las letras. En Go, %d: la o con tilde son 2 bytes.\n", len(s))

	fmt.Println()
	fmt.Println("== 5.2 · LA TRAMPA: indexar da un byte, no una letra ==")
	fmt.Printf("  s[11] -> %d (%q)   <- el primer byte de la o con tilde: medio caracter\n", s[11], string(s[11]))
	fmt.Println()
	fmt.Println("== 5.3 · range recorre letras (runas), no bytes ==")
	for i, r := range "añb" {
		fmt.Printf("  posicion de byte %d -> %q\n", i, r)
	}
	fmt.Println("  Fijate: salta del 1 al 3. La ene ocupa dos bytes.")
}
