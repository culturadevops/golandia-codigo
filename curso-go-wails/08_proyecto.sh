#!/usr/bin/env bash
# Lección 8 · El proyecto, en dos escenarios. Se lanza DESDE LA MÁQUINA (no dentro del contenedor):
# monta una red con nubecita y abre el inspector en el contenedor de Wails.
#   1. nubecita viva, y a los 5 s otro proceso mete 12 mensajes en «pedidos»
#   2. nubecita apagada
set -u
export MSYS_NO_PATHCONV=1
D=${CURSO_HOST:-$(cd "$(dirname "$0")" && pwd -W)}
RED=curso-wails-red
docker rm -f nubecita-curso >/dev/null 2>&1; docker network rm $RED >/dev/null 2>&1
docker network create $RED >/dev/null
docker run -d --name nubecita-curso --network $RED nubecita:curso >/dev/null
caja() { docker run --rm --network $RED -e ENDPOINT=http://nubecita-curso:4566 -e NOMBRE="$1" \
           -v gomod-wails:/go/pkg/mod -v gocache-wails:/root/.cache/go-build -v "$D:/app" curso-wails:1 bash -c "$2"; }

echo "== 8.1 · nubecita viva, y llegan mensajes con la ventana abierta =="
caja 08_proyecto-vivo "go run ./08_proyecto/sembrar && (sleep 6; go run ./08_proyecto/sembrar -mas 12) & bash capturar.sh 08_proyecto 10" \
  | grep -v "go: downloading"

echo
echo "== 8.2 · nubecita apagada =="
docker stop nubecita-curso >/dev/null
caja 08_proyecto-sin-servidor "bash capturar.sh 08_proyecto 6" | grep -v "go: downloading"
docker rm -f nubecita-curso >/dev/null 2>&1; docker network rm $RED >/dev/null 2>&1
