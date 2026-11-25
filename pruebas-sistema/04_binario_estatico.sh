#!/bin/sh
# 4 · El binario estatico: el argumento de verdad de por que Go se comio la infraestructura.
# Se ejecuta DENTRO del contenedor de Go (ver correr.sh).
set -e

cd /pruebas

cat > servidor.go <<'FIN'
package main

import (
	"fmt"
	"net/http"
)

func main() {
	http.HandleFunc("/salud", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "vivo")
	})
	fmt.Println("escuchando en el 8080")
	http.ListenAndServe(":8080", nil)
}
FIN

echo "== a) compilar para esta maquina =="
go build -o servidor servidor.go
ls -l servidor | awk '{printf "   servidor: %s bytes\n", $5}'

echo
echo "== b) que necesita para arrancar? =="
# ldd sobre un binario de Go sin cgo responde que no depende de NADA.
ldd servidor 2>&1 | head -3 || true
file servidor 2>/dev/null | cut -c1-120 || true

echo
echo "== c) compilar para OTRA maquina, sin salir de esta =="
echo "   (un Raspberry Pi de 64 bits, y un servidor Windows, desde Linux x86)"
GOOS=linux  GOARCH=arm64 go build -o servidor-arm64     servidor.go
GOOS=windows GOARCH=amd64 go build -o servidor.exe      servidor.go
GOOS=darwin GOARCH=arm64 go build -o servidor-mac       servidor.go
ls -l servidor-arm64 servidor.exe servidor-mac | awk '{printf "   %-18s %s bytes\n", $9, $5}'

echo
echo "== d) cuantos destinos sabe hacer este compilador, sin instalar nada =="
go tool dist list | wc -l
