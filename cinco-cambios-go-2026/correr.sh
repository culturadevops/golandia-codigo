#!/bin/bash
# Golandia 714 · corre dentro de golang:1.27.2. Copia las demos a un directorio temporal (mktemp):
# /trabajo se monta de solo lectura y go fix reescribe archivos solo dentro de la copia.
set -u
export GOFLAGS=-mod=mod GOTOOLCHAIN=local
T=$(mktemp -d)
cp -r /trabajo/demos "$T/"
cd "$T/demos" || exit 1

run() { echo "\$ $*"; "$@" 2>&1 | sed "s#$T/demos#…#g"; echo "[código de salida: ${PIPESTATUS[0]}]"; }

echo "go version: $(go version)"

echo; echo "## 01-go-fix-servicio · go.mod go 1.27"
cd 01-go-fix-servicio
printf 'module pedidos\n\ngo 1.27\n' > go.mod
run go build ./...
run go vet ./...
run go test ./...
run go fix -diff ./...
go fix -diff ./... 2>/dev/null > /tmp/diff127.txt
echo "# líneas quitadas / añadidas por go fix: $(grep -c '^-[^-]' /tmp/diff127.txt) / $(grep -c '^+[^+]' /tmp/diff127.txt)"
echo "### JSON-127"; go vet -vettool="$(go env GOTOOLDIR)/fix" -json ./... 2>&1; echo "### FIN-JSON"
run go fix ./...
run go build ./...
run go vet ./...
run go test ./...
run go fix -diff ./...
cd ..

echo; echo "## 01-go-fix-servicio · el mismo código con go.mod go 1.21"
rm -rf servicio121 && cp -r /trabajo/demos/01-go-fix-servicio servicio121 && cd servicio121
printf 'module pedidos\n\ngo 1.21\n' > go.mod
run go test ./...
go fix -diff ./... 2>/dev/null > /tmp/diff121.txt
echo "# líneas quitadas / añadidas por go fix: $(grep -c '^-[^-]' /tmp/diff121.txt) / $(grep -c '^+[^+]' /tmp/diff121.txt)"
run go fix -diff ./...
echo "### JSON-121"; go vet -vettool="$(go env GOTOOLDIR)/fix" -json ./... 2>&1; echo "### FIN-JSON"
cd ..

echo; echo "## 02-go-fix-inline · //go:fix inline"
cd 02-go-fix-inline
printf 'module demo\n\ngo 1.27\n' > go.mod
run cat viejo/viejo.go
run go run ./app
run go fix -diff ./...
run go fix ./...
run cat app/main.go
run go run ./app
cd ..

echo; echo "## 03-go-tool-fix-help · cuántos analizadores"
echo "\$ go tool fix help | grep -cE '^    [a-z]+ '"
go tool fix help 2>&1 | grep -cE '^    [a-z]+ '
rm -rf "$T"
