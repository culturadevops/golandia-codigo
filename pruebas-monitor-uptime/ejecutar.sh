#!/bin/bash
# Corre el banco del monitor de uptime en golang:1.27.2 y deja la salida literal en SALIDA.md. Desde Git Bash:
#   bash canales/golandia/recursos/pruebas-monitor-uptime/ejecutar.sh
# Todo es local: los sitios y el «Discord» son servidores en 127.0.0.1 dentro del contenedor. Nada sale a Internet.
cd "$(dirname "$0")" || exit 1
export MSYS_NO_PATHCONV=1
AQUI=$(cygpath -w "$PWD")
{
  echo "# SALIDA — monitor de uptime en Go (golandia)"
  echo
  echo "> Corrida del $(date '+%Y-%m-%d %H:%M'). golang:1.27.2, solo librería estándar, contenedor con --cpus 2. Tiempos redondeados a 100 ms."
  echo "> El «discord» es un servidor falso local que imita el webhook (POST {\"content\": ...} → 204). No se usa ningún webhook real."
  echo
  echo '```'
  docker run --rm --cpus 2 --network none -v gomodcache:/go/pkg/mod -v "$AQUI:/trabajo" golang:1.27.2 bash /trabajo/correr.sh 2>&1
  echo '```'
} > SALIDA.md
cat SALIDA.md
