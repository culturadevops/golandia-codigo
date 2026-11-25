#!/bin/bash
# Corre la comparación goroutines vs hilos en golang:1.27.2 con 2 CPU y deja la salida literal en SALIDA.md. Desde Git Bash:
#   bash canales/golandia/recursos/pruebas-goroutines-vs-hilos/ejecutar.sh
cd "$(dirname "$0")" || exit 1
export MSYS_NO_PATHCONV=1
AQUI=$(cygpath -w "$PWD")
{
  echo "# SALIDA — Goroutines vs hilos (golandia)"
  echo
  echo "> Corrida del $(date '+%Y-%m-%d %H:%M'). golang:1.27.2, contenedor con --cpus 2 --memory 4g. Cada línea es la corrida mediana de 5."
  echo
  echo '```'
  docker run --rm --cpus 2 --memory 4g -v "$AQUI:/trabajo" golang:1.27.2 bash /trabajo/correr.sh 2>&1
  echo '```'
} > SALIDA.md
cat SALIDA.md
