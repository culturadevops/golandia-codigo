#!/bin/bash
ORIGEN="$1"
DESTINO="$2"
ARCHIVOS=$(find "$ORIGEN" -name "*.log" -type f -mtime +7)
for ARCHIVO in $ARCHIVOS; do
    gzip "$ARCHIVO"
    mv "$ARCHIVO.gz" "$DESTINO"
done
