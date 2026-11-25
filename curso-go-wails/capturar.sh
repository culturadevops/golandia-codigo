#!/usr/bin/env bash
# capturar.sh DIR SEGUNDOS [CLICS]  — dentro del contenedor curso-wails:1
# Compila la lección DIR con las etiquetas de Wails, la abre en una pantalla virtual (Xvfb),
# espera, hace los clics que se le pidan (x,y separados por espacios), la fotografía y enseña lo
# que la app escribió. La captura va a /app/capturas/DIR.png.
set -u
DIR=$1; ESPERA=${2:-6}; CLICS=${3:-}
TAGS="desktop,production,webkit2_41"
go build -tags "$TAGS" -o /tmp/app "./$DIR" || { echo "  NO COMPILA"; exit 1; }
Xvfb :99 -screen 0 900x640x24 >/dev/null 2>&1 &
sleep 1
DISPLAY=:99 /tmp/app >/tmp/app.log 2>&1 &
APP=$!
sleep "$ESPERA"
for c in $CLICS; do DISPLAY=:99 xdotool mousemove ${c%,*} ${c#*,} click 1; sleep 1.5; done
if kill -0 $APP 2>/dev/null; then
  DISPLAY=:99 import -window root "/app/capturas/${NOMBRE:-$DIR}.png"
  echo "  ventana viva · captura: capturas/${NOMBRE:-$DIR}.png"
  sleep "${DESPUES:-0}"
  kill -0 $APP 2>/dev/null || { wait $APP; echo "  ⛔ después de la captura, la app terminó sola (código $?)"; }
else
  wait $APP; echo "  ⛔ la app terminó sola (código $?)"
fi
kill $APP 2>/dev/null; wait $APP 2>/dev/null
grep -v -E "^(Gtk-Message|\(app)" /tmp/app.log | grep -v '^$' | sed 's/^/  app> /'
exit 0
