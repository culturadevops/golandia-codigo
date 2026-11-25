#!/usr/bin/env bash
# Lección 7 — salida para personas y -json para scripts, leída con jq.
source "$(dirname "$0")/../comun.sh"
compilar vigia ./vigia
compilar ingenuo07 ./07_json/ingenuo
echo "jq: $(jq --version)"

seccion "7.1 · La misma revisión, para personas y para scripts"
ej 'vigia revisar $SRV/ok "$SRV/lento?ms=250" $SRV/roto 2>/dev/null'
ej 'vigia revisar -json $SRV/ok "$SRV/lento?ms=250" $SRV/roto 2>/dev/null'

seccion "7.2 · jq: las URLs caídas, y la más lenta"
ej 'vigia revisar -json $SRV/ok "$SRV/lento?ms=250" $SRV/roto 2>/dev/null | jq -r ".[] | select(.ok | not) | .url"'
ej 'vigia revisar -json $SRV/ok "$SRV/lento?ms=250" $SRV/roto 2>/dev/null | jq -r "max_by(.latencia_ms) | \"\(.url) \(.latencia_ms) ms\""'

seccion "7.3 · Ingenuo: time.Duration en JSON son nanosegundos"
ej 'ingenuo07 "$SRV/lento?ms=250"'
ej 'ingenuo07 "$SRV/lento?ms=250" | jq ".[0].latencia > 1000"'

seccion "7.4 · Ingenuo: sin caídos, el JSON es null y no []"
ej 'ingenuo07 -solo-caidos $SRV/ok'
ejc 'ingenuo07 -solo-caidos $SRV/ok | jq -r ".[].url"'
ej 'vigia revisar -json -solo-caidos $SRV/ok 2>/dev/null'
ejc 'vigia revisar -json -solo-caidos $SRV/ok 2>/dev/null | jq -r ".[].url"'

seccion "7.5 · Ingenuo: un aviso por stdout rompe el JSON entero"
ejc 'ingenuo07 -avisar $SRV/ok "$SRV/lento?ms=250" | jq length'
ej 'vigia revisar -json $SRV/ok "$SRV/lento?ms=250" | jq length'
