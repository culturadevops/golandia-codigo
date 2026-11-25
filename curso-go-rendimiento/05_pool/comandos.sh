#!/bin/sh
# Lección 5: sync.Pool, y cuándo no sirve.
cd "$(dirname "$0")/.." && . ./comun.sh

echo '## El servicio entero: io.ReadAll (V3) contra búferes del pool (V5)'
paso "go test ./resumen -run '^\$' -bench Handler -benchmem -count 10 > \$T/handler.txt"
paso "go tool benchstat -col /version \$T/handler.txt"

echo '## Un objeto pequeño'
paso "go test ./05_pool -run '^\$' -bench . -benchmem -count 10 > \$T/pool.txt"
paso "go tool benchstat -filter '.name:Pequeno' -col /con \$T/pool.txt"
paso "go tool benchstat -filter '.name:Grande' -col /con \$T/pool.txt"

echo '## El pool no es una caché'
paso "go test ./05_pool -run TestSobreviveAlGC -v -count 1"
