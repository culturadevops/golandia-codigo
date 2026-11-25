#!/bin/bash

ORIGEN="$1"
DESTINO="$2"

if [ ! -d "$ORIGEN" ]; then
  echo "El directorio de origen no existe."
  exit 1
fi

if [ ! -d "$DESTINO" ]; then
  echo "El directorio de destino no existe."
  exit 1
fi

find "$ORIGEN" -type f -name "*.log" -mtime +7 -print0 | while read -d '' -r archivo; do
  nombre_archivo=$(basename "$archivo")
  gzip "$archivo"
  mv "$archivo.gz" "$DESTINO/$nombre_archivo.gz"
done
