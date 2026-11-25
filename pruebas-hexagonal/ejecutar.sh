#!/bin/bash
# Ejecuta la demo de arquitectura hexagonal y deja la salida literal en SALIDA.md. Desde Git Bash en Windows:
#   bash canales/golandia/recursos/pruebas-hexagonal/ejecutar.sh
cd "$(dirname "$0")" || exit 1
export MSYS_NO_PATHCONV=1
AQUI=$(cygpath -w "$PWD")
docker rm -f pg-hex >/dev/null 2>&1
docker network create hexagonal >/dev/null 2>&1
docker run -d --name pg-hex --network hexagonal --network-alias pg -e POSTGRES_USER=tienda -e POSTGRES_PASSWORD=tienda \
  -e POSTGRES_DB=tienda postgres:17 >/dev/null
until docker exec pg-hex pg_isready -U tienda >/dev/null 2>&1; do sleep 1; done
sleep 2
docker exec pg-hex psql -U tienda -c "CREATE TABLE pedidos (id bigserial PRIMARY KEY, cliente text NOT NULL, unidades int NOT NULL, precio int NOT NULL)" >/dev/null
{
  echo "# SALIDA — Arquitectura hexagonal en Go (golandia)"
  echo
  echo "> Corrida del $(date '+%Y-%m-%d %H:%M'). Salida literal de correr.sh dentro de golang:1.26, con Postgres 17."
  echo
  echo '```'
  docker run --rm --network hexagonal -v "$AQUI:/trabajo" golang:1.26 bash -c "go version; bash /trabajo/correr.sh" 2>&1
  echo '```'
} > SALIDA.md
docker rm -f pg-hex >/dev/null 2>&1
cat SALIDA.md
