#!/bin/sh
# Lección 7: GOGC y GOMEMLIMIT, medidos. Memoria pico contra CPU.
cd "$(dirname "$0")/.." && . ./comun.sh

paso "go build -o \$T/gc ./07_gc"

echo '## GOGC: tres ejecuciones de cada valor'
for g in 25 50 100 200 400; do
	paso "GOGC=$g \$T/gc; GOGC=$g \$T/gc; GOGC=$g \$T/gc"
done

echo '## GOMEMLIMIT'
paso "GOGC=off GOMEMLIMIT=200MiB \$T/gc; GOGC=off GOMEMLIMIT=200MiB \$T/gc; GOGC=off GOMEMLIMIT=200MiB \$T/gc"
paso "GOMEMLIMIT=100MiB \$T/gc; GOMEMLIMIT=100MiB \$T/gc; GOMEMLIMIT=100MiB \$T/gc"
paso "GOMEMLIMIT=60MiB \$T/gc; GOMEMLIMIT=60MiB \$T/gc; GOMEMLIMIT=60MiB \$T/gc"

echo '## gctrace: una línea por ciclo del recolector'
paso "GODEBUG=gctrace=1 \$T/gc 2>&1 | grep '^gc ' | sed -n '9,12p'"
