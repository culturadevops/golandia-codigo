#!/bin/sh
cp -r /d/demos/caidas /tmp/c && cd /tmp/c
echo "### gleam run"; gleam run > /tmp/o 2>&1; c=$?; grep -vE '^\s*(Downloading|Downloaded|Resolving|Added|Compiling|Compiled|Running)' /tmp/o; echo "[código de salida: $c]"
