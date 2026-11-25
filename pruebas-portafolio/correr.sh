#!/bin/bash
# Corre DENTRO de golang:1.26 (lo lanza ejecutar.sh). El mismo escáner, antes y después de dejarlo «de portafolio».
set -u
rm -rf /w && mkdir -p /w && cp -r /trabajo/antes /trabajo/despues /w/
echo "### ANTES: el escáner del vídeo 239, tal cual"
cd /w/antes
echo "\$ find . -name '*.go' | xargs wc -l"
find . -name '*.go' | xargs wc -l
echo "\$ go test ./..."
go test ./... 2>&1
echo "\$ go vet ./..."
go vet ./... 2>&1 && echo "(sin avisos)"
echo
echo "### DESPUÉS: lógica en un paquete, pruebas, contexto"
cd /w/despues
printf 'module escaner\n\ngo 1.26\n' > go.mod
echo "\$ find . -name '*.go' | xargs wc -l"
find . -name '*.go' | sort | xargs wc -l
echo "\$ go vet ./..."
go vet ./... 2>&1 && echo "(sin avisos)"
echo "\$ go test -race -cover ./..."
go test -race -cover ./... 2>&1
echo "\$ go test -race -v -run 'Test' ./escaner | grep -E '^(=== RUN|--- )'"
go test -race -v ./escaner 2>&1 | grep -E "^--- |^    --- "
echo
echo "\$ go build -o /tmp/escaner ./cmd/escaner && /tmp/escaner -hasta 100"
go build -o /tmp/escaner ./cmd/escaner && /tmp/escaner -hasta 100
