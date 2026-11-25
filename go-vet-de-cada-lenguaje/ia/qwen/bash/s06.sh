#!/bin/bash
ORIGEN="$1"
DESTINO="$2"
TIEMPO=$((7*24*60*60))

# Comprobar si los directorios de origen y destino existen
if [ ! -d "$ORIGEN" ]; then
    echo "El directorio de origen no existe."
    exit 1
fi

if [ ! -d "$DESTINO" ]; then
    echo "El directorio de destino no existe."
    exit 1
fi

# Comprimir y mover archivos .log modificados hace más de 7 días
find "$ORIGEN" -name "*.log" -type f -mtime +7 -exec gzip {} \; -exec mv {}.gz "$DESTINO" \;

echo "Proceso completado."
