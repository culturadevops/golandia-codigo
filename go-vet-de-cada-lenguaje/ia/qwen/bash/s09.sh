#!/bin/bash

ORIGEN="$1"
DESTINO="$2"

if [[ ! -d "$ORIGEN" || ! -d "$DESTINO" ]]; then
    echo "Uso: $0 <directorio_origen> <directorio_destino>"
    exit 1
fi

find "$ORIGEN" -name "*.log" -type f -mtime +7 -exec gzip {} \;
find "$ORIGEN" -name "*.log.gz" -type f -exec mv {} "$DESTINO" \;

echo "Proceso completado. Los archivos .log.gz han sido movidos al directorio de destino."
