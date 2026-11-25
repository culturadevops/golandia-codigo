#!/usr/bin/env bash
# sondeo.sh — ¿compilan de verdad los dos, y qué hace falta instalar para cada uno?
# ⛔ Antes de escribir una línea de guion. Regla del banco desde el vídeo 229.
#
# ⚠️ TRES CORRECCIONES DE LA PRIMERA CORRIDA, las tres mías:
#  1. Git Bash convierte `-w /app` en `C:/Program Files/Git/app`; docker falla y el script
#     salía con 0. "Pasó" sin compilar nada. → MSYS_NO_PATHCONV=1.
#  2. Fyne NO compila con lo que dice su documentación (libgl1-mesa-dev + xorg-dev):
#     glfw pide wayland-client-core.h. → hay que añadir las cabeceras de Wayland.
#  3. Wails 2.15 exige Go >= 1.25. Con golang:1.24 ni resuelve el módulo.
# ⛔ Y LA MISMA IMAGEN PARA LOS DOS LADOS, o la comparación no vale.
set -u
export MSYS_NO_PATHCONV=1
D=$(cd "$(dirname "$0")" && pwd -W)
IMG=golang:1.25-bookworm

echo "############ FYNE ############"
docker run --rm -v "$D/app-fyne:/app" -w /app $IMG bash -c '
set -e
apt-get update -qq >/dev/null 2>&1
apt-get install -y -qq libgl1-mesa-dev xorg-dev \
  libwayland-dev libxkbcommon-dev wayland-protocols >/tmp/apt.log 2>&1 \
  || { echo "APT FALLO"; tail -5 /tmp/apt.log; exit 1; }
echo "apt: ok"
go get fyne.io/fyne/v2@latest 2>&1 | tail -2
go mod tidy 2>&1 | tail -2
go build -o /tmp/contador . && echo "BUILD: OK" || { echo "BUILD: FALLO"; exit 1; }
ls -l /tmp/contador
'
echo "  >>> fyne salio con $?"

echo
echo "############ WAILS ############"
docker run --rm -v "$D/app-wails:/app" -w /app $IMG bash -c '
set -e
apt-get update -qq >/dev/null 2>&1
apt-get install -y -qq libgtk-3-dev libwebkit2gtk-4.1-dev >/tmp/apt.log 2>&1 \
  || { echo "APT FALLO"; tail -5 /tmp/apt.log; exit 1; }
echo "apt: ok"
go get github.com/wailsapp/wails/v2@latest 2>&1 | tail -2
go mod tidy 2>&1 | tail -2
go build -tags webkit2_41 -o /tmp/contador . && echo "BUILD: OK" || { echo "BUILD: FALLO"; exit 1; }
ls -l /tmp/contador
'
echo "  >>> wails salio con $?"
