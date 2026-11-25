#!/bin/sh
set -u
cd /app
export CGO_ENABLED=1
echo "############ e08 · tema propio y recursos empotrados ############"
if ! go build -o /tmp/e08 ./etapas/e08-tema 2>/tmp/err.txt; then
    echo "[X] NO COMPILA:"; head -6 /tmp/err.txt; exit 1
fi
echo "  el icono en disco:         $(stat -c%s etapas/e08-tema/icono.png) bytes"
echo "  binario CON el icono:      $(stat -c%s /tmp/e08) bytes"
echo "  binario de la ventana e01: $(go build -o /tmp/e01 ./etapas/e01-ventana && stat -c%s /tmp/e01) bytes"
Xvfb :99 -screen 0 2560x1600x24 >/dev/null 2>&1 &
sleep 2
export DISPLAY=:99 FYNE_SCALE=3
/tmp/e08 >/tmp/salida-e08.txt 2>&1 &
PID=$!
sleep 5
import -window root -crop 1920x900+0+0 +repage /app/capturas/e08-tema.png 2>/dev/null
sleep 0.3
kill $PID 2>/dev/null; wait $PID 2>/dev/null
echo "############ FIN ############"
