#!/bin/sh
# F · El punto en que SQLite deja de valer: MAS DE UN PROCESO contra el mismo fichero.
# Dentro de un proceso, el pool de Go serializa a los escritores en memoria y sale barato.
# Entre procesos no hay pool compartido: la unica coordinacion es el candado del fichero.
set -u
mkdir -p /datos
D="file:/datos/f.db?_pragma=journal_mode(WAL)&_pragma=busy_timeout(5000)&_pragma=synchronous(NORMAL)"
S=${SEGUNDOS:-4}

echo "############ F · UN PROCESO CONTRA DOS, LA MISMA CARGA TOTAL ############"
echo "-- 1 proceso con 8 escritores --"
./banco -motor sqlite -segundos $S -escritores 8 -maxabiertas 1 -dsn "$D" \
  -etiqueta "1 proceso  · 8 escritores"

echo "-- 2 procesos con 4 escritores cada uno, a la vez --"
./banco -motor sqlite -segundos 1 -escritores 1 -maxabiertas 1 -dsn "$D" -etiqueta "(preparando)" >/dev/null
./banco -motor sqlite -segundos $S -escritores 4 -maxabiertas 1 -dsn "$D" -limpiar=false \
  -etiqueta "2 procesos · proceso A (4 escritores)" &
A=$!
./banco -motor sqlite -segundos $S -escritores 4 -maxabiertas 1 -dsn "$D" -limpiar=false \
  -etiqueta "2 procesos · proceso B (4 escritores)" &
B=$!
wait $A $B
echo "############ FIN ############"
