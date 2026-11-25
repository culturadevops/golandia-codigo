#!/bin/bash
# Respalda los .log de un directorio en otro. Uso: respaldo_bien.sh ORIGEN DESTINO
set -euo pipefail
origen=$(realpath "$1")
destino=$2

cd "$destino" || exit 1
cp "$origen"/*.log .
for f in ./*.log; do
  gzip "$f"
done
echo "respaldo listo en $(pwd)"
