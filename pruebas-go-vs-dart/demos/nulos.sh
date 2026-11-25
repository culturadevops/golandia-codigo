#!/bin/sh
cp -r /d/demos/nulos /tmp/n && cd /tmp/n
echo "### dart analyze"; dart analyze > /tmp/o 2>&1; c=$?; cat /tmp/o; echo "[código de salida: $c]"
echo "### dart compile exe bin/nulos.dart"; dart compile exe bin/nulos.dart -o /tmp/nulos > /tmp/o 2>&1; c=$?; cat /tmp/o; echo "[código de salida: $c]"
