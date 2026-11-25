#!/bin/bash

ORIGEN="$1"
DESTINO="$2"

if [ -z "$ORIGEN" ] || [ -z "$DESTINO" ]; then
    echo "Uso: $0 <directorio_de_origen> <directorio_de_destino>"
    exit 1
fi

find "$ORIGEN" -name "*.log" -type f -mtime +7 -print0 | while IFS= read -r -d $'\0' log_file; do
    compressed_file="${log_file}.gz"
    mv "$log_file" "$compressed_file"
    mv "$compressed_file" "$DESTINO"
done
