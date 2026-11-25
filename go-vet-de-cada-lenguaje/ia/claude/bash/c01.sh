#!/usr/bin/env bash
set -euo pipefail

if [[ $# -ne 2 ]]; then
    echo "Uso: $0 <directorio_origen> <directorio_destino>" >&2
    exit 1
fi

origen="$1"
destino="$2"

if [[ ! -d "$origen" ]]; then
    echo "Error: el directorio de origen '$origen' no existe." >&2
    exit 1
fi

mkdir -p "$destino"

find "$origen" -maxdepth 1 -type f -name '*.log' -mtime +7 -print0 |
while IFS= read -r -d '' archivo; do
    nombre="$(basename "$archivo")"
    salida="$destino/$nombre.gz"

    if [[ -e "$salida" ]]; then
        echo "Aviso: '$salida' ya existe, se omite '$archivo'." >&2
        continue
    fi

    if gzip -c -- "$archivo" > "$salida"; then
        touch -r "$archivo" "$salida"
        rm -f -- "$archivo"
        echo "Comprimido y movido: $archivo -> $salida"
    else
        echo "Error al comprimir '$archivo'." >&2
        rm -f -- "$salida"
    fi
done
