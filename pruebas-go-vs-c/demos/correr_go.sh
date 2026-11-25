#!/bin/sh
# La versión en Go de los mismos tres errores. Se ejecuta dentro de golang:1.27.2 con /d montado.
cp -r /d/demos/go /tmp/g && cd /tmp/g
echo "### go vet ./..."; go vet ./... 2>&1; echo "[código de salida: $?]"
echo "### go run ."; go run . 2>&1; echo "[código de salida: $?]"
