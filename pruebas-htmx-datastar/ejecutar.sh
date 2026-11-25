#!/bin/bash
# Ejecuta la demo de HTMX contra Datastar y deja la salida literal en SALIDA.md. Desde Git Bash en Windows:
#   bash canales/golandia/recursos/pruebas-htmx-datastar/ejecutar.sh
cd "$(dirname "$0")" || exit 1
export MSYS_NO_PATHCONV=1
AQUI=$(cygpath -w "$PWD")
docker rm -f demo-hd >/dev/null 2>&1
{
  echo "# SALIDA — ¿HTMX o Datastar? Con Go (golandia)"
  echo
  echo "> Corrida del $(date '+%Y-%m-%d %H:%M'). htmx 2.0.11 · Datastar 1.0.4 · datastar-go v1.2.2 · golang:1.26."
  echo
  echo '```'
  docker run --rm -v "$AQUI:/trabajo" golang:1.26 bash -c "go version; bash /trabajo/correr.sh" 2>&1
  echo '```'
  echo
  echo "### 4 · En un navegador de verdad (Chromium, Playwright): 5 s quieto, escribe «ru», 2 s"
  echo
  echo '```'
  docker run -d --name demo-hd -p 8090:8080 -v "$AQUI:/trabajo" golang:1.26 bash -c \
    "rm -rf /w && mkdir -p /w && cp -r /trabajo/app /w/app && cd /w/app && mkdir -p static && cp /trabajo/htmx.min.js /trabajo/datastar.js static/ && go mod init demo >/dev/null 2>&1 && go get github.com/starfederation/datastar-go@v1.2.2 >/dev/null 2>&1 && go mod tidy >/dev/null 2>&1 && go run ." >/dev/null
  until curl -s -m 2 http://127.0.0.1:8090/htmx > /dev/null; do sleep 1; done
  node navegador.js 8090 2>&1
  echo '```'
} > SALIDA.md
docker rm -f demo-hd >/dev/null 2>&1
cat SALIDA.md
