#!/bin/sh
# Corredor del curso MIEMBROS de Fyne (idea #83). Compila cada etapa, la arranca sobre un
# servidor grafico de mentira (Xvfb) y captura la ventana. Son ventanas de verdad, no maquetas.
#
# ⚠️ Fyne 2.8 exige cabeceras de Wayland ademas de las de X11. Sin libwayland-dev,
#    wayland-protocols y libxkbcommon-dev el build muere con
#    "fatal error: wayland-client-core.h: No such file or directory".
set -u
cd /app
export CGO_ENABLED=1
t() { echo; echo "############ $1 ############"; }

t "ENTORNO"
go version
go list -m fyne.io/fyne/v2

Xvfb :99 -screen 0 2560x1600x24 >/dev/null 2>&1 &
sleep 2
export DISPLAY=:99
# 🥇 FYNE_SCALE=3: sin esto la ventana sale de ~430 px y el motor de video la estira a 1920,
# y se ve borrosa. Con escala 3 la captura ya nace grande y entra nitida. (2026-09-03)
export FYNE_SCALE=3

captura() { # $1 = carpeta de la etapa   $2 = nombre   $3 = espera   $4 = recorte
  echo; echo "-- $2 --"
  if ! go build -o /tmp/$2 "./$1" 2>/tmp/err.txt; then
      echo "[X] NO COMPILA:"; head -5 /tmp/err.txt; return 0
  fi
  echo "  binario: $(stat -c%s /tmp/$2) bytes"
  /tmp/$2 >/tmp/salida-$2.txt 2>&1 &
  PID=$!
  sleep "$3"
  import -window root -crop "$4" +repage "/app/capturas/$2.png" 2>/dev/null
  sleep 0.3
  kill $PID 2>/dev/null; wait $PID 2>/dev/null
  L=$(grep -vE "XDG_RUNTIME_DIR|dbus-launch|failed to watch theme|Cause:|At: " /tmp/salida-$2.txt | head -8)
  if [ -n "$L" ]; then echo "  lo que imprimio:"; echo "$L" | sed 's/^/    /'; else echo "  (no imprimio nada)"; fi
}

t "e01 · la primera ventana"
captura etapas/e01-ventana e01-ventana 3 1920x660+0+0

t "e02 · contenedores: quien decide el tamano"
captura etapas/e02-contenedores e02-contenedores 3 1920x900+0+0

t "e03 · binding: se toca el DATO, no la etiqueta"
captura etapas/e03-binding e03-binding 3 1920x780+0+0

t "e04 · MAL: tocar la interfaz desde una gorrutina"
captura etapas/e04-mal e04-mal 3 1920x600+0+0

t "e04 · BIEN: el retoque dentro de fyne.Do"
captura etapas/e04-bien e04-bien 3 1920x600+0+0

t "e05 · 50.000 objetos, y cuantas filas construye de verdad"
captura etapas/e05-lista e05-lista 4 1920x1080+0+0

t "e07 · formulario con validacion"
captura etapas/e07-formulario e07-formulario 3 1920x780+0+0

t "FIN"
