#!/bin/sh
# Lección 6: go tool trace. Goroutines bloqueadas y latencia del planificador.
cd "$(dirname "$0")/.." && . ./comun.sh

echo '## El mismo trabajo, con el candado en tres sitios'
paso "go build -o \$T/traza ./06_trace"
paso "\$T/traza -modo ancho -repetir 7"
paso "\$T/traza -modo estrecho -repetir 7"
paso "\$T/traza -modo local -repetir 7"

echo '## La traza del candado ancho'
paso "\$T/traza -modo ancho -trace \$T/ancho.out"
paso "go tool trace -pprof=sync \$T/ancho.out > \$T/sync.prof && go tool pprof -top -nodecount=6 \$T/sync.prof"
paso "go tool trace -pprof=sched \$T/ancho.out > \$T/sched.prof && go tool pprof -top -nodecount=5 \$T/sched.prof"
paso "python3 06_trace/desglose.py \$T/ancho.out"

echo '## La traza con un mapa por goroutine'
paso "\$T/traza -modo local -trace \$T/local.out"
paso "python3 06_trace/desglose.py \$T/local.out"
