#!/bin/sh
set -u
cd /app
export CGO_ENABLED=1
echo "############ e06 · el cliente conectado a nubecita ############"
go get github.com/aws/aws-sdk-go-v2/service/s3 github.com/aws/aws-sdk-go-v2/credentials 2>&1 | tail -2
go mod tidy 2>&1 | tail -1
if ! go build -o /tmp/e06 ./etapas/e06-conectado 2>/tmp/err.txt; then
    echo "[X] NO COMPILA:"; head -6 /tmp/err.txt; exit 1
fi
echo "  binario: $(stat -c%s /tmp/e06) bytes"
Xvfb :99 -screen 0 2560x1600x24 >/dev/null 2>&1 &
sleep 2
export DISPLAY=:99 FYNE_SCALE=3
/tmp/e06 >/tmp/salida-e06.txt 2>&1 &
PID=$!
sleep 6
import -window root -crop 1920x1080+0+0 +repage /app/capturas/e06-conectado.png 2>/dev/null
sleep 0.3
kill $PID 2>/dev/null; wait $PID 2>/dev/null
echo "  lo que imprimio:"
grep -vE "XDG_RUNTIME_DIR|dbus-launch|failed to watch theme|Cause:|At: " /tmp/salida-e06.txt | head -6
echo "############ FIN ############"
