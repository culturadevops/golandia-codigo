#!/bin/sh
cp -r /d/demos/go /tmp/g && cd /tmp/g
echo "### go build ."; go build . 2>&1; echo "[código de salida: $?]"
