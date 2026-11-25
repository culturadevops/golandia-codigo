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

if [[ "$(realpath "$origen")" == "$(realpath "$destino")" ]]; then
    echo "Error: origen y destino no pueden ser el mismo directorio." >&2
    exit 1
fi

procesados=0
errores=0

while IFS= read -r -d '' archivo; do
    nombre="$(basename "$archivo")"
    salida="$destino/$nombre.gz"

    if [[ -e "$salida" ]]; then
        echo "Aviso: '$salida' ya existe, se omite '$archivo'." >&2
        errores=$((errores + 1))
        continue
    fi

    if gzip -c -- "$archivo" > "$salida"; then
        touch -r "$archivo" -- "$salida"
        rm -f -- "$archivo"
        echo "Comprimido y movido: $archivo -> $salida"
        procesados=$((procesados + 1))
    else
        echo "Error al comprimir '$archivo'." >&2
        rm -f -- "$salida"
        errores=$((errores + 1))
    fi
done < <(find "$origen" -maxdepth 1 -type f -name '*.log' -mtime +7 -print0)

echo "Procesados: $procesados | Omitidos/errores: $errores"
[[ $errores -eq 0 ]]
