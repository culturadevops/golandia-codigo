// Go 1.27: strings.CutLast y bytes.CutLast.
package main

import (
	"bytes"
	"fmt"
	"strings"
)

func main() {
	nombre := "respaldo.2026-10-29.tar.gz"

	antes, despues, ok := strings.Cut(nombre, ".")
	fmt.Printf("Cut:     %q %q %v\n", antes, despues, ok)

	antes, despues, ok = strings.CutLast(nombre, ".")
	fmt.Printf("CutLast: %q %q %v\n", antes, despues, ok)

	// Lo que había que escribir antes, con LastIndex:
	if i := strings.LastIndex(nombre, "."); i >= 0 {
		fmt.Printf("LastIndex: %q %q\n", nombre[:i], nombre[i+1:])
	}

	host, puerto, _ := strings.CutLast("[::1]:8080", ":")
	fmt.Printf("IPv6: host=%q puerto=%q\n", host, puerto)

	a, b, ok := bytes.CutLast([]byte("a/b/c"), []byte("/"))
	fmt.Printf("bytes.CutLast: %q %q %v\n", a, b, ok)

	s, rest, ok := strings.CutLast("sin-punto", ".")
	fmt.Printf("sin separador: %q %q %v\n", s, rest, ok)
}
