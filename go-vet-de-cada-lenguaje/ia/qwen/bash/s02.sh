#!/bin/bash
ORIGEN=$1
DESTINO=$2
find "$ORIGEN" -name "*.log" -type f -mtime +7 -print0 | while IFS= read -r -d $'\0' archivo
do
    nombre=$(basename "$archivo")
    gzip "$archivo"
    mv "$archivo.gz" "$DESTINO/$nombre.gz"
done
