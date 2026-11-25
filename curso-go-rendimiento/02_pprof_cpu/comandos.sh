#!/bin/sh
# Lección 2: perfil de CPU con pprof (top, top -cum, list, peek).
cd "$(dirname "$0")/.." && . ./comun.sh

echo '## Perfil de CPU de la versión 1'
paso "go test ./resumen -run '^\$' -bench 'Resumen/version=V1\$' -benchtime 3s -cpuprofile \$T/cpu_v1.prof -o \$T/resumen.test > /dev/null"
paso "go tool pprof -top -nodecount=10 \$T/resumen.test \$T/cpu_v1.prof"
paso "go tool pprof -top -cum -nodecount=14 \$T/resumen.test \$T/cpu_v1.prof"
paso "go tool pprof -list 'resumen.ResumenV1\$' \$T/resumen.test \$T/cpu_v1.prof"
paso "go tool pprof -peek 'regexp.MustCompile\$' \$T/resumen.test \$T/cpu_v1.prof"

echo '## Versión 2: la expresión regular, compilada una vez'
paso "go test ./resumen -run '^\$' -bench 'Resumen/version=V2\$' -benchtime 3s -cpuprofile \$T/cpu_v2.prof -o \$T/resumen.test > /dev/null"
paso "go tool pprof -top -cum -nodecount=10 \$T/resumen.test \$T/cpu_v2.prof"
bench_resumen
paso "go tool benchstat -filter '/version:(V1 OR V2)' -col /version \$T/resumen.txt"
