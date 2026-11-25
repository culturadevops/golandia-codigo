#!/usr/bin/env bash
# Lección 3 — la misma CLI con Cobra, y lo que cuesta: tamaño del binario y arranque, medidos.
source "$(dirname "$0")/../comun.sh"
compilar vigia02 ./02_subcomandos
compilar vigia03 ./03_cobra
compilar medir ./medir

seccion "3.1 · Lo mismo que la lección 2"
ejc 'vigia03 revisar --timeout 1s $SRV/ok $SRV/roto'

seccion "3.2 · Con Cobra (pflag) los flags van donde quieras"
ejc 'vigia03 revisar $SRV/ok -t 1s'

seccion "3.3 · Un subcomando mal escrito: Cobra sugiere"
ejc 'vigia03 revsar $SRV/ok'

seccion "3.4 · Ojo: en pflag un guion es un flag corto; -timeout es -t con el valor «imeout»"
ej 'vigia03 revisar -timeout 1s $SRV/ok 2>&1 | head -1'
ejc 'vigia03 revisar -timeout 1s $SRV/ok > /dev/null 2>&1'

seccion "3.5 · Tamaño del binario (bytes)"
compilar t02 ./02_subcomandos
compilar t03 ./03_cobra
compilar t02-sw ./02_subcomandos -trimpath -ldflags=-s\ -w
compilar t03-sw ./03_cobra -trimpath -ldflags=-s\ -w
for b in t02 t03 t02-sw t03-sw; do
	printf '  %-8s %10s bytes\n' "$b" "$(stat -c %s "$BIN/$b")"
done

seccion "3.6 · Paquetes que entran en el binario"
printf '  %-14s %4s paquetes\n' "flag (02)" "$(go list -deps ./02_subcomandos | wc -l)"
printf '  %-14s %4s paquetes\n' "cobra (03)" "$(go list -deps ./03_cobra | wc -l)"

seccion "3.7 · Arranque: «version» lanzado 300 veces"
echo "  flag : $(medir -n 300 "$BIN/t02-sw" version)"
echo "  cobra: $(medir -n 300 "$BIN/t03-sw" version)"
echo "  flag : $(medir -n 300 "$BIN/t02-sw" version)   (repetido, para ver el ruido)"
echo "  cobra: $(medir -n 300 "$BIN/t03-sw" version)   (repetido, para ver el ruido)"

seccion "3.8 · Líneas de código de cada versión"
ej 'grep -c "" 02_subcomandos/main.go 03_cobra/main.go'
