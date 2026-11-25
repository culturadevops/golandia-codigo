#!/bin/sh
cp -r /d/demos/go_caidas /tmp/g && cd /tmp/g && go build -o /tmp/caidas . 2>&1
echo "### ./caidas"; /tmp/caidas > /tmp/o 2>&1; c=$?; grep -E 'terminé|panic:|sigo vivo' /tmp/o; echo "[código de salida: $c]"
