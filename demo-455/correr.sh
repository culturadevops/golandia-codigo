#!/usr/bin/env bash
# Demo del vídeo 455 «¿Debo aprender Go en la era de la IA?» — se EJECUTA y su salida va a SALIDA.md (regla: la terminal del
# vídeo se ejecuta). Dos cosas que el guion afirma:
#   1) una variable sin usar no compila (lo típico que deja suelto un código generado);
#   2) se compila para otro sistema desde esta máquina y sale UN binario.
# Trabaja dentro de un directorio temporal: no toca nada fuera.
set -u
cd "$(dirname "$0")"
T=$(mktemp -d)
cd "$T"
go mod init demo/precios >/dev/null 2>&1

cat > main.go <<'EOF'
package main

import (
	"fmt"
	"os"
	"strconv"
)

func main() {
	total := 0
	precio, err := strconv.Atoi(os.Args[1])
	if err != nil {
		fmt.Println("precio inválido:", err)
		os.Exit(1)
	}
	fmt.Println("con impuesto:", precio*118/100)
}
EOF

echo '$ cat main.go'
cat main.go
echo
echo '$ go build -o precios .'
go build -o precios . 2>&1
echo "(código de salida: $?)"
echo
# el arreglo: quitar la variable que nadie usa
sed -i '/total := 0/d' main.go
echo '$ go build -o precios . && ./precios 100'
go build -o precios . 2>&1 && ./precios 100
echo
echo '$ GOOS=linux GOARCH=amd64 go build -o precios-linux .'
GOOS=linux GOARCH=amd64 go build -o precios-linux . 2>&1
echo '$ ls -l precios-linux | awk "{print \$5, \$9}"'
ls -l precios-linux | awk '{print $5, $9}'
echo '$ file precios-linux'
file precios-linux 2>/dev/null | cut -d, -f1-2
cd / && rm -rf "$T"
