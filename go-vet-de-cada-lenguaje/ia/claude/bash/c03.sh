#!/usr/bin/env bash
# Uso: ./archivar_logs.sh <directorio_origen> <directorio_destino>
# Comprime con gzip los archivos .log del origen modificados hace más de 7 días
# y mueve los .gz resultantes al destino.

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

if [[ ! -w "$destino" ]]; then
    echo "Error: no hay permisos de escritura en '$destino'." >&2
    exit 1
fi

procesados=0
errores=0

while IFS= read -r -d '' archivo; do
    nombre="$(basename "$archivo")"
    gz="${archivo}.gz"

    if [[ -e "$destino/${nombre}.gz" ]]; then
        echo "Aviso: '$destino/${nombre}.gz' ya existe, se omite '$archivo'." >&2
        errores=$((errores + 1))
        continue
    fi

    if gzip -- "$archivo" && mv -- "$gz" "$destino/"; then
        echo "Archivado: $archivo -> $destino/${nombre}.gz"
        procesados=$((procesados + 1))
    else
        echo "Error al procesar '$archivo'." >&2
        errores=$((errores + 1))
    fi
done < <(find "$origen" -maxdepth 1 -type f -name '*.log' -mtime +7 -print0)

echo "Completado: $procesados archivo(s) archivado(s), $errores con errores u omitidos."

[[ $errores -eq 0 ]]
