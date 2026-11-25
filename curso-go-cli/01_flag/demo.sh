#!/usr/bin/env bash
# Lección 1 — el paquete flag: lo que hace bien y las dos trampas que no dan error.
source "$(dirname "$0")/../comun.sh"
compilar vigia01 ./01_flag
TIMEFORMAT='[tardó %1R s]'

seccion "1.1 · Lo normal: flags primero, URLs después"
ej 'time vigia01 -timeout 1s $SRV/ok $SRV/lento'

seccion "1.2 · Trampa: un flag DESPUÉS de una URL ya no es un flag"
ej 'time vigia01 -v $SRV/lento -timeout 1s'

seccion "1.3 · Trampa: un flag booleano no se lleva el valor que va detrás"
ej 'vigia01 -v false $SRV/ok'
ej 'vigia01 -v=false $SRV/ok'

seccion "1.4 · Lo que flag te da gratis: -h, valores con tipo y el código 2"
ejc 'vigia01 -h'
ejc 'vigia01 -timeout 5 $SRV/ok'
ejc 'vigia01 --timeout=1s $SRV/ok'
