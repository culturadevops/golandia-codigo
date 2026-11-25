#!/bin/sh
# Lección 8: la condición de carrera que las pruebas no ven sin -race.
cd "$(dirname "$0")/.." && . ./comun.sh

echo '## Las pruebas de siempre'
paso "go test ./resumen -count 1"
paso "go test ./08_carrera -run TestHandlerV4Concurrente -count 200"

echo '## Con el detector de carreras'
paso "go test -race ./08_carrera -run TestHandlerV4Concurrente -count 1 2>&1 | head -45"
paso "n=0; for i in 1 2 3 4 5 6 7 8 9 10; do go test -race ./08_carrera -run TestHandlerV4Concurrente -count 1 > /dev/null 2>&1 || n=\$((n+1)); done; echo \"carrera detectada en \$n de 10 ejecuciones\""

echo '## Y con tráfico de verdad, sin -race'
paso "go run ./10_carga -version V4 -clientes 32 -duracion 5s"

echo '## El arreglo: escribir antes de devolver el búfer'
paso "go test -race ./08_carrera -run TestHandlerV5Concurrente -count 10"
