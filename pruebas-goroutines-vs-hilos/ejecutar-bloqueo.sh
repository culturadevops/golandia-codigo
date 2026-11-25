#!/bin/bash
# Corre la prueba de goroutines bloqueadas en llamadas al sistema y deja la salida en SALIDA-bloqueo.md. Desde Git Bash:
#   bash canales/golandia/recursos/pruebas-goroutines-vs-hilos/ejecutar-bloqueo.sh
cd "$(dirname "$0")" || exit 1
export MSYS_NO_PATHCONV=1
AQUI=$(cygpath -w "$PWD")
{
  echo "# SALIDA — Goroutines bloqueadas en una llamada al sistema (golandia)"
  echo
  echo "> Corrida del $(date '+%Y-%m-%d %H:%M'). golang:1.27.2, contenedor con --cpus 2 --memory 4g. Hilos contados a 1 s de una espera de 2 s."
  echo
  echo '```'
  docker run --rm --cpus 2 --memory 4g -v "$AQUI:/trabajo" golang:1.27.2 bash /trabajo/correr-bloqueo.sh 2>&1
  echo '```'
} > SALIDA-bloqueo.md
cat SALIDA-bloqueo.md
