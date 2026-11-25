#!/usr/bin/env bash
# Lección 4 — configuración con precedencia: flags > entorno > archivo > por defecto.
source "$(dirname "$0")/../comun.sh"
compilar vigia ./vigia
compilar ingenuo04 ./04_config/ingenuo
TMP="$(mktemp -d)"
cd "$TMP" || exit 1
unset VIGIA_TIMEOUT VIGIA_CONCURRENCIA VIGIA_CONFIG

seccion "4.1 · Sin nada: valores por defecto"
ej 'vigia config'
ej 'ingenuo04'

seccion "4.2 · Con archivo vigia.json"
echo '{"timeout": "5s", "concurrencia": 8}' > vigia.json
ej 'cat vigia.json'
ej 'vigia config'
ej 'ingenuo04'

seccion "4.3 · Archivo + entorno"
ej 'VIGIA_TIMEOUT=3s vigia config'
ej 'VIGIA_TIMEOUT=3s ingenuo04'

seccion "4.4 · Archivo + entorno + flag"
ej 'VIGIA_TIMEOUT=3s vigia config -timeout 1s'
ej 'VIGIA_TIMEOUT=3s ingenuo04 -timeout 1s'

seccion "4.5 · El flag puesto a mano con el MISMO valor que el defecto"
ej 'VIGIA_TIMEOUT=3s vigia config -timeout 10s'
ej 'VIGIA_TIMEOUT=3s ingenuo04 -timeout 10s'

seccion "4.6 · Un valor mal escrito en el entorno (falta la unidad)"
ejc 'VIGIA_TIMEOUT=3 vigia config'
ejc 'VIGIA_TIMEOUT=3 ingenuo04'

seccion "4.7 · El archivo por defecto es opcional; el que pides, no"
rm vigia.json
ejc 'vigia config'
ejc 'vigia config -config /no/existe.json'

cd / && rm -rf "$TMP"
