#!/bin/sh
# Corre dentro de koalaman/shellcheck-alpine:v0.11.0. /w = la carpeta del vídeo, montada de solo lectura.
shellcheck --version | head -2
cd /w/casos/bash
for f in respaldo.sh respaldo_bien.sh; do
  echo "=== caso|$f|shellcheck"; shellcheck -f gcc "$f" 2>&1; echo "--- exit=$?"
done
for f in /w/ia/*/bash/*.sh; do
  cd "$(dirname "$f")"
  echo "=== ia|${f#/w/}|shellcheck"; shellcheck -f gcc "$(basename "$f")" 2>&1; echo "--- exit=$?"
done
