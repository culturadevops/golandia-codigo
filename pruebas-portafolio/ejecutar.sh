#!/bin/bash
# Ejecuta la demo de «5 proyectos para tu portafolio» y deja la salida literal en SALIDA.md. Desde Git Bash:
#   bash canales/golandia/recursos/pruebas-portafolio/ejecutar.sh
cd "$(dirname "$0")" || exit 1
export MSYS_NO_PATHCONV=1
AQUI=$(cygpath -w "$PWD")
{
  echo "# SALIDA — Un proyecto de portafolio, antes y después (golandia)"
  echo
  echo "> Corrida del $(date '+%Y-%m-%d %H:%M'). golang:1.26, golangci-lint v2.15.0, Docker."
  echo
  echo '```'
  docker run --rm -v "$AQUI:/trabajo" golang:1.26 bash -c "go version; bash /trabajo/correr.sh" 2>&1
  echo
  echo "### golangci-lint (v2.15.0), antes y después"
  echo "\$ golangci-lint run ./...   # antes"
  docker run --rm -v "$AQUI\\antes:/app" -w /app golangci/golangci-lint:v2.15.0 golangci-lint run ./... 2>&1 | tail -8
  echo "\$ golangci-lint run ./...   # después"
  docker run --rm -v "$AQUI\\despues:/app" -w /app golangci/golangci-lint:v2.15.0 golangci-lint run ./... 2>&1 | tail -8
  echo
  echo "### La imagen"
  echo "\$ docker build -t escaner . && docker image ls escaner"
  printf 'module escaner\n\ngo 1.26\n' > despues/go.mod
  docker build -q -t golandia-portafolio-escaner:1 despues >/dev/null 2>&1
  docker image inspect -f '{{.Size}}' golandia-portafolio-escaner:1 | awk '{printf "%.1f MiB (descarga comprimida)\n", $1/1048576}'
  echo "\$ docker run --rm escaner -hasta 10"
  docker run --rm golandia-portafolio-escaner:1 -hasta 10 2>&1
  echo '```'
} > SALIDA.md
cat SALIDA.md
