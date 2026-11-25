#!/bin/sh
cp -r /d/demos/go_nulos /tmp/g && cd /tmp/g
echo "### go vet ."; go vet . 2>&1; echo "[código de salida: $?]"
echo "### go build . && ./demo"; go build -o demo . 2>&1; ./demo > /tmp/o 2>&1; c=$?; head -2 /tmp/o; echo "[código de salida: $c]"
