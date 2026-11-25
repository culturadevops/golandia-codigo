#!/usr/bin/env bash
# Demo del vídeo 464 «os.Root en Go» — se EJECUTA y su salida va a SALIDA.md.
# Un servidor de archivos de mentira: el usuario pide un nombre y se lee de ./publico.
#   A) filepath.Join + os.Open  (lo de siempre);
#   B) os.OpenRoot + root.Open  (Go 1.24+);
#   C) filepath.IsLocal como comprobación previa (Go 1.20+).
# Fuera de ./publico hay un secreto.txt de mentira. Todo dentro de un directorio temporal.
set -u
T=$(mktemp -d)
cd "$T"
go mod init demo/archivos >/dev/null 2>&1
mkdir -p publico/docs
echo "hola, soy público" > publico/leeme.txt
echo "manual en docs" > publico/docs/manual.txt
echo "CLAVE_FALSA_DE_DEMO=1234" > secreto.txt

cat > main.go <<'EOF'
package main

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

func leer(f io.Reader) string {
	b, _ := io.ReadAll(f)
	return strings.TrimSpace(string(b))
}

func main() {
	nombre := os.Args[1]
	fmt.Printf("pide: %q\n", nombre)

	// A) lo de siempre
	f, err := os.Open(filepath.Join("publico", nombre))
	if err != nil {
		fmt.Println("  A) os.Open:      error:", err)
	} else {
		fmt.Println("  A) os.Open:      lee:", leer(f))
		f.Close()
	}

	// B) os.Root: nada fuera de ./publico
	root, _ := os.OpenRoot("publico")
	defer root.Close()
	g, err := root.Open(nombre)
	if err != nil {
		fmt.Println("  B) root.Open:    error:", err)
	} else {
		fmt.Println("  B) root.Open:    lee:", leer(g))
		g.Close()
	}

	// C) comprobación previa
	fmt.Println("  C) IsLocal:     ", filepath.IsLocal(nombre))
}
EOF

go build -o archivos . 2>&1
echo '$ go version'
go version | cut -d' ' -f3
echo '$ ls . publico'
ls . publico | grep -v -E '^(archivos|go.mod|main.go)$'
echo
for n in leeme.txt docs/manual.txt ../secreto.txt docs/../../secreto.txt; do
  echo "\$ ./archivos $n"
  ./archivos "$n"
  echo
done
cd / && rm -rf "$T"
