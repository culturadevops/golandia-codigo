#!/bin/bash
# Corre dentro de python:3.13.18-slim (trae bash). TODO pasa dentro de un mktemp: nada fuera se toca.
T=$(mktemp -d)
cd "$T" || exit 1
mkdir "mis logs" "mis respaldos"
echo uno > "mis logs/app 1.log"
echo dos > "mis logs/web.log"
for s in respaldo.sh respaldo_bien.sh; do
  echo "=== caso|$s|ejecuta"
  echo "\$ bash $s \"mis logs\" \"mis respaldos\""
  bash "/w/casos/bash/$s" "mis logs" "mis respaldos" 2>&1 | sed "s#$T#/tmp/demo#g"
  echo "--- exit=${PIPESTATUS[0]}"
  echo "\$ find . -type f | sort"; find . -type f | sort
done
