#!/bin/sh
# El mismo programa con go run y con tinygo run: ¿compila? ¿hace lo mismo?
cp -r /d/demos/prueba /tmp/p && cp -r /d/demos/json /tmp/p/json
for n in json plantilla exec metodo plugin; do
  cd /tmp/p/$n
  echo "### $n · go run ."; go run . > /tmp/o 2>&1; c=$?; head -3 /tmp/o; echo "[código de salida: $c]"
  echo "### $n · tinygo run ."; tinygo run -no-debug . > /tmp/o 2>&1; c=$?; head -4 /tmp/o; echo "[código de salida: $c]"
done
