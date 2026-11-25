#!/bin/sh
set -e

echo "== go version =="
go version

echo
echo "== instalar delve v1.27.2 =="
go install github.com/go-delve/delve/cmd/dlv@v1.27.2 >/tmp/install.log 2>&1 || (cat /tmp/install.log; exit 1)
GOBIN="$(go env GOPATH)/bin"
"$GOBIN/dlv" version

cd /lab/proyecto
go mod tidy >/tmp/tidy.log 2>&1 || (cat /tmp/tidy.log; exit 1)

echo
echo "== el programa, corrido normal, sin debugger (se corta con timeout) =="
timeout 3 go run . 2>&1 || echo "(se colgo: exactamente lo que se espera de un deadlock parcial)"

echo
echo "== ahora con dlv: breakpoint justo despues de que las goroutines se traban =="
"$GOBIN/dlv" debug --allow-non-terminal-interactive=true main.go < comandos.txt 2>&1 | sed -E 's/\x1b\[[0-9;]*m//g'
