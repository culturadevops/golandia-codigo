#!/bin/sh
cp -r /d/demos/isolates /tmp/i && cd /tmp/i
echo "### dart run bin/isolates.dart"; dart run bin/isolates.dart 2>&1; echo "[código de salida: $?]"
