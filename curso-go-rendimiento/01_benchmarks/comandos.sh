#!/bin/sh
# Lección 1: benchmarks bien hechos (b.Loop, -benchmem, -count, benchstat).
cd "$(dirname "$0")/.." && . ./comun.sh

echo '## La trampa: un benchmark que no mide nada'
paso "go test ./01_benchmarks -bench . -benchmem -count 5"

echo '## Antes de medir: todas las versiones dan la misma respuesta'
paso "go test ./resumen -run TestTodasIguales -v -count 1"

echo '## Pocas repeticiones: benchstat no puede afirmar nada'
paso "go test ./resumen -run '^\$' -bench Resumen -benchmem -count 3 > \$T/n3.txt"
paso "go tool benchstat -col /version \$T/n3.txt"

echo '## Diez repeticiones'
bench_resumen
paso "grep 'version=V1' \$T/resumen.txt | head -3"
paso "go tool benchstat -col /version \$T/resumen.txt"
