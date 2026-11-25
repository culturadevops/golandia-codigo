#!/bin/sh
cp -r /d/demos/tipos /tmp/t && cd /tmp/t
echo "### gleam build"; gleam build > /tmp/o 2>&1; c=$?; grep -vE '^\s*(Downloading|Downloaded|Resolving|Added|Compiling|Compiled)' /tmp/o; echo "[código de salida: $c]"
sed -i 's/    Enviado -> "en camino"/    Enviado -> "en camino"\n    Devuelto -> "devuelto"/' src/tipos.gleam
echo "### (agrego el caso Devuelto) gleam run"; gleam run > /tmp/o 2>&1; c=$?; grep -vE '^\s*(Downloading|Downloaded|Resolving|Added|Compiling|Compiled|Running)' /tmp/o; echo "[código de salida: $c]"
