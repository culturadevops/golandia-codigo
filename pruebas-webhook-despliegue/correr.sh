#!/bin/bash
# Corre la demo en un contenedor desechable: bash correr.sh > SALIDA-crudo.txt
cd "$(dirname "$0")" || exit 1
MSYS_NO_PATHCONV=1 docker run --rm --cpus 2 -v "$(pwd -W 2>/dev/null || pwd)":/pruebas golang:1.27.2 bash /pruebas/ejecutar.sh
