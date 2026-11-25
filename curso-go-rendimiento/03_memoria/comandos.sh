#!/bin/sh
# Lección 3: perfil de memoria. alloc_space (lo que se asignó) contra inuse_space (lo que sigue vivo).
cd "$(dirname "$0")/.." && . ./comun.sh

echo '## alloc_space en el benchmark de la versión 2'
paso "go test ./resumen -run '^\$' -bench 'Resumen/version=V2\$' -benchtime 2s -memprofile \$T/mem_v2.prof -o \$T/resumen.test > /dev/null"
paso "go tool pprof -sample_index=alloc_space -top -nodecount=8 \$T/resumen.test \$T/mem_v2.prof"
paso "go tool pprof -sample_index=alloc_space -list 'resumen.ResumenV2\$' \$T/resumen.test \$T/mem_v2.prof"

echo '## El servicio con el panel de errores: 300 peticiones'
paso "go build -o \$T/memoria ./03_memoria && \$T/memoria -perfil \$T/heap.prof"
paso "go tool pprof -sample_index=alloc_space -top -nodecount=6 \$T/memoria \$T/heap.prof"
paso "go tool pprof -sample_index=inuse_space -top -nodecount=5 \$T/memoria \$T/heap.prof"
paso "go tool pprof -sample_index=inuse_space -list 'main.guardarErrores' \$T/memoria \$T/heap.prof"

echo '## Con strings.Clone'
paso "\$T/memoria -perfil \$T/heap2.prof -arreglado"
paso "go tool pprof -sample_index=inuse_space -top -nodecount=5 \$T/memoria \$T/heap2.prof"
