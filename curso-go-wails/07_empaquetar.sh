#!/usr/bin/env bash
# Lección 7 · Empaquetar. Dentro del contenedor curso-wails:1.
# El fallo del vídeo 232: `go build` a secas sobre un proyecto de Wails compila, termina con código
# 0... y el binario no abre ninguna ventana. Hacen falta las etiquetas desktop y production.
set -u
cd /app
mide() { printf '  %-44s %12s bytes · %3s librerías del sistema\n' "$1" "$(stat -c %s "$2")" "$(ldd "$2" 2>/dev/null | wc -l)"; }

echo "== 7.1 · go build a secas =="
go build -o /tmp/a-secas ./01_ventana; echo "  go build: código $?"
mide "sin etiquetas" /tmp/a-secas
Xvfb :98 -screen 0 800x600x24 >/dev/null 2>&1 & sleep 1
DISPLAY=:98 timeout 8 /tmp/a-secas >/tmp/s.log 2>&1; echo "  al ejecutarlo: código $? en menos de 8 s"
head -3 /tmp/s.log | sed 's/^/  app> /'

echo
echo "== 7.2 · con las etiquetas de Wails =="
go build -tags desktop,production,webkit2_41 -o /tmp/con ./01_ventana
mide "-tags desktop,production" /tmp/con
DISPLAY=:98 timeout 8 /tmp/con >/dev/null 2>&1; c=$?
[ $c = 124 ] && echo "  al ejecutarlo: sigue abierta a los 8 s (la cerró timeout)" || echo "  al ejecutarlo: código $c"
go build -tags desktop,production,webkit2_41 -ldflags="-s -w" -o /tmp/con-sw ./01_ventana
mide "  + -ldflags=\"-s -w\"" /tmp/con-sw

echo
echo "== 7.3 · Para Windows, desde Linux =="
GOOS=windows GOARCH=amd64 CGO_ENABLED=0 go build -tags desktop,production -ldflags="-s -w -H windowsgui" -o /tmp/app.exe ./01_ventana
echo "  GOOS=windows: código $? · $(stat -c %s /tmp/app.exe) bytes"
echo "  (en Windows usa WebView2, el motor de Edge: aquí no se puede abrir, y no se dice que se vea igual)"
