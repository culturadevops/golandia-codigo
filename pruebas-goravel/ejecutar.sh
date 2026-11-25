#!/bin/bash
# Ejecuta la demo completa de Goravel y deja la salida literal en SALIDA.md. Desde Git Bash en Windows:
#   bash canales/golandia/recursos/pruebas-goravel/ejecutar.sh
cd "$(dirname "$0")" || exit 1
export MSYS_NO_PATHCONV=1
AQUI=$(cygpath -w "$PWD")
docker rm -f pg >/dev/null 2>&1
docker network create goravel >/dev/null 2>&1
docker run -d --name pg --network goravel -e POSTGRES_USER=goravel -e POSTGRES_PASSWORD=goravel -e POSTGRES_DB=tienda postgres:17 >/dev/null
until docker exec pg pg_isready -U goravel >/dev/null 2>&1; do sleep 1; done
sleep 2
{
  echo "# SALIDA — Goravel: ¿el Laravel de Go? (golandia)"
  echo
  echo "> Corrida del $(date '+%Y-%m-%d %H:%M'). Salida literal de correr.sh + parte2.sh dentro de golang:1.26, con Postgres 17."
  echo
  echo '```'
  docker run --rm --network goravel -v "$AQUI:/trabajo" golang:1.26 bash -c "go version; bash /trabajo/correr.sh; bash /trabajo/parte2.sh" 2>&1 \
    | sed 's/\x1b\[[0-9;]*m//g'
  echo '```'
} > SALIDA.md
docker rm -f pg >/dev/null 2>&1
tail -50 SALIDA.md
