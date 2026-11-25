#!/bin/sh
cp -r /d/demos/go /tmp/g && cd /tmp/g
echo "### go vet ."; go vet . 2>&1; echo "[código de salida: $?]"
echo "### go run ."; go run . 2>&1; echo "[código de salida: $?]"
