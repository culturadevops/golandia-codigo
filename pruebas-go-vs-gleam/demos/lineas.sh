#!/bin/sh
# Líneas de código de cada servidor: sin líneas vacías ni comentarios de línea entera.
for f in /srvgo/main.go /d/servidor/src/servidor.gleam; do
  n=$(grep -v -E '^\s*$|^\s*(//|#|/\*|\*)' "$f" | wc -l)
  echo "$n líneas  $(echo "$f" | sed 's#^/srvgo/#go/#; s#^/d/##')"
done
