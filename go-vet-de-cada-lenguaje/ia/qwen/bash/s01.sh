#!/bin/bash

ORIGEN="$1"
DESTINO="$2"

# Comprueba si los argumentos son válidos
if [ -z "$ORIGEN" ] || [ -z "$DESTINO" ]; then
  echo "Uso: $0 <directorio_origen> <directorio_destino>"
  exit 1
fi

# Comprueba si el directorio origen existe
if [ ! -d "$ORIGEN" ]; then
  echo "El directorio de origen no existe."
  exit 1
fi

# Comprueba si el directorio destino existe, si no, lo crea
if [ ! -d "$DESTINO" ]; then
  mkdir -p "$DESTINO"
fi

# Comprime los archivos .log modificados hace más de 7 días y los mueve al destino
find "$ORIGEN" -type f -name "*.log" -mtime +7 -print0 | xargs -0 -I {} sh -c 'gzip "{}" && mv "{}.gz" "$DESTINO"'

echo "Proceso completado."
