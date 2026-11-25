#!/bin/bash
# Corre las demos de io.Reader en golang:1.27.2 y deja la salida literal en SALIDA.md. Desde Git Bash:
#   bash canales/golandia/recursos/pruebas-io-reader/ejecutar.sh
cd "$(dirname "$0")" || exit 1
export MSYS_NO_PATHCONV=1
AQUI=$(cygpath -w "$PWD")
{
  echo "# SALIDA — io.Reader en Go (golandia)"
  echo
  echo "> Corrida del $(date '+%Y-%m-%d %H:%M'). golang:1.27.2, contenedor con --cpus 2 --memory 4g. Archivo de prueba: 500 MB de texto generado."
  echo
  echo '```'
  docker run --rm --cpus 2 --memory 4g -v "$AQUI:/trabajo" golang:1.27.2 bash /trabajo/correr.sh 2>&1
  echo '```'
} > SALIDA.md
cat SALIDA.md
