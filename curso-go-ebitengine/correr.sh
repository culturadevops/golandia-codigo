#!/usr/bin/env bash
# correr.sh DIR [NOMBRE] — dentro del contenedor curso-ebiten:1.
# Compila la lección, la abre en Xvfb y, cuando el juego lo pide (crea /tmp/foto-NOMBRE), hace la
# captura en capturas/NOMBRE.png. Enseña lo que imprime el juego.
set -u
DIR=$1; NOMBRE=${2:-$1}
go build -o /tmp/juego "./$DIR" || { echo "  NO COMPILA"; exit 1; }
Xvfb :99 -screen 0 960x540x24 >/dev/null 2>&1 &
sleep 1
rm -f /tmp/foto-*
DISPLAY=:99 /tmp/juego > /tmp/juego.log 2>&1 &
J=$!
[ -x "/app/$DIR/pulsar.sh" ] || [ -f "/app/$DIR/pulsar.sh" ] && bash "/app/$DIR/pulsar.sh" &
while kill -0 $J 2>/dev/null; do
  for f in /tmp/foto-*; do
    [ -e "$f" ] || continue
    n=${f#/tmp/foto-}; DISPLAY=:99 import -window root "/app/capturas/${NOMBRE:-$n}.png"; rm -f "$f"
    echo "  captura: capturas/${NOMBRE:-$n}.png"
  done
  sleep 0.05
done
wait $J; c=$?
grep -v -E "^ALSA|^$" /tmp/juego.log
[ $c = 0 ] || echo "  código de salida $c"
exit 0
