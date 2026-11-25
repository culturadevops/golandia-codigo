#!/bin/sh
# Lección 4: análisis de escape (-gcflags=-m) y cómo bajar las asignaciones.
cd "$(dirname "$0")/.." && . ./comun.sh

echo '## Lo que escapa en la versión 2'
paso "go build -gcflags=-m ./resumen 2>&1 | grep 'v2.go' | grep -vE 'inlin|does not escape|leaking'"

echo '## Tres casos aislados'
paso "go test -gcflags=-m ./04_escape 2>&1 | grep 'escape_test.go' | grep -E 'moved to heap|metodo escapes|ruta escapes'"
paso "go test ./04_escape -run '^\$' -bench . -benchmem -count 10 > \$T/escape.txt"
paso "go tool benchstat -filter '.name:Decodificar' -col /e \$T/escape.txt"
paso "go tool benchstat -filter '.name:Clave' -col /forma \$T/escape.txt"

echo '## La versión 3'
paso "go build -gcflags=-m ./resumen 2>&1 | grep 'v3.go' | grep -E 'moved to heap|string\(clave\)|make\(map'"
bench_resumen
paso "go tool benchstat -filter '/version:(V2 OR V3)' -col /version \$T/resumen.txt"
