#!/bin/bash
# Repite ingenuo y estados 5 veces para ver cuánto varían los números (el borde del segundo 3 cae a un lado u otro).
#   bash canales/golandia/recursos/pruebas-monitor-uptime/repetir.sh  → SALIDA-repeticiones.md
cd "$(dirname "$0")" || exit 1
export MSYS_NO_PATHCONV=1
AQUI=$(cygpath -w "$PWD")
{
  echo "# SALIDA — repeticiones del monitor (golandia)"
  echo
  echo "> Corrida del $(date '+%Y-%m-%d %H:%M'). golang:1.27.2, --cpus 2, --network none. 5 repeticiones de cada modo."
  echo
  echo '```'
  docker run --rm --cpus 2 --network none -v gomodcache:/go/pkg/mod -v "$AQUI:/trabajo" golang:1.27.2 bash -c '
    cd /trabajo/monitor && go build -o /tmp/monitor . &&
    for i in 1 2 3 4 5; do echo "ingenuo #$i: $(/tmp/monitor ingenuo | grep avisos)"; done &&
    for i in 1 2 3 4 5; do echo "estados #$i: $(/tmp/monitor estados | grep -E "CAÍDO|volvió|avisos" | tr -s " " | tr "\n" "|")"; done' 2>&1
  echo '```'
} > SALIDA-repeticiones.md
cat SALIDA-repeticiones.md
