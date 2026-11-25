#!/bin/bash
# Corre las diferencias y el banco de encoding/json vs encoding/json/v2 en golang:1.27.2 y deja la salida en SALIDA.md. Desde Git Bash:
#   bash canales/golandia/recursos/pruebas-json-v2/ejecutar.sh
cd "$(dirname "$0")" || exit 1
export MSYS_NO_PATHCONV=1
AQUI=$(cygpath -w "$PWD")
{
  echo "# SALIDA — encoding/json/v2 en Go 1.27 (golandia)"
  echo
  echo "> Corrida del $(date '+%Y-%m-%d %H:%M'). golang:1.27.2, contenedor con --cpus 2. Banco: 10.000 objetos (~1,6 MB), mediana de 5 pasadas. Notas de Go 1.27: fuente-go1.27.html."
  echo
  echo '```'
  docker run --rm --cpus 2 -e GOFLAGS=-buildvcs=false -v "$AQUI:/trabajo" golang:1.27.2 bash /trabajo/correr.sh 2>&1
  echo '```'
} > SALIDA.md
cat SALIDA.md
