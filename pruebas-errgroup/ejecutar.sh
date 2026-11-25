#!/bin/bash
# Corre las demos de errgroup en golang:1.27.2 y deja la salida literal en SALIDA.md. Desde Git Bash:
#   bash canales/golandia/recursos/pruebas-errgroup/ejecutar.sh
cd "$(dirname "$0")" || exit 1
export MSYS_NO_PATHCONV=1
AQUI=$(cygpath -w "$PWD")
{
  echo "# SALIDA — errgroup en Go (golandia)"
  echo
  echo "> Corrida del $(date '+%Y-%m-%d %H:%M'). golang:1.27.2 + golang.org/x/sync v0.24.0, contenedor con --cpus 2. Tiempos redondeados a 10 ms."
  echo
  echo '```'
  docker run --rm --cpus 2 -v gomodcache:/go/pkg/mod -v "$AQUI:/trabajo" golang:1.27.2 bash /trabajo/correr.sh 2>&1
  echo '```'
} > SALIDA.md
cat SALIDA.md
