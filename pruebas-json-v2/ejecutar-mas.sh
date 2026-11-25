#!/bin/bash
# Más diferencias v1/v2 (mas/main.go) en golang:1.27.2; deja la salida literal en SALIDA-mas.md. Desde Git Bash:
#   bash canales/golandia/recursos/pruebas-json-v2/ejecutar-mas.sh
cd "$(dirname "$0")" || exit 1
export MSYS_NO_PATHCONV=1
AQUI=$(cygpath -w "$PWD")
{
  echo "# SALIDA — más diferencias entre encoding/json y encoding/json/v2 (golandia)"
  echo
  echo "> Corrida del $(date '+%Y-%m-%d %H:%M'). golang:1.27.2."
  echo
  echo '```'
  echo '$ go run ./mas'
  docker run --rm -e GOFLAGS=-buildvcs=false -v "$AQUI:/trabajo" -w /trabajo/mas golang:1.27.2 go run . 2>&1
  echo '```'
} > SALIDA-mas.md
cat SALIDA-mas.md
