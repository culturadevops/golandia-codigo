#!/bin/sh
cp -r /d/demos/go_comparte /tmp/g && cd /tmp/g
echo "### go run ."; go run . 2>&1; echo "[código de salida: $?]"
echo "### go run -race ."; go run -race . > /tmp/o 2>&1; c=$?; grep -E 'WARNING|contador|Found' /tmp/o; echo "[código de salida: $c]"
