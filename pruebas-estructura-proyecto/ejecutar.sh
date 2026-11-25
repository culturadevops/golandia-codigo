#!/bin/bash
# Compila las cuatro estructuras de proyecto en golang:1.27.2 y deja la salida literal en SALIDA.md. Desde Git Bash:
#   bash canales/golandia/recursos/pruebas-estructura-proyecto/ejecutar.sh
cd "$(dirname "$0")" || exit 1
export MSYS_NO_PATHCONV=1
AQUI=$(cygpath -w "$PWD")
{
  echo "# SALIDA — Estructura de un proyecto Go (golandia)"
  echo
  echo "> Corrida del $(date '+%Y-%m-%d %H:%M'). golang:1.27.2. Fuentes: go.dev/doc/modules/layout (fuente-go.dev-modules-layout.html) y el README de golang-standards/project-layout (fuente-project-layout-README.md)."
  echo
  echo '```'
  docker run --rm -e GOFLAGS=-buildvcs=false -v "$AQUI:/trabajo" golang:1.27.2 bash /trabajo/correr.sh 2>&1
  echo '```'
} > SALIDA.md
cat SALIDA.md
