#!/usr/bin/env bash
# Lección 2 — subcomandos con la biblioteca estándar: un FlagSet por subcomando.
source "$(dirname "$0")/../comun.sh"
compilar vigia02 ./02_subcomandos

seccion "2.1 · Cada subcomando con sus flags"
ejc 'vigia02 revisar -timeout 1s $SRV/ok $SRV/roto'
ejc 'vigia02 version'

seccion "2.2 · Un flag delante del subcomando no es de nadie"
ejc 'vigia02 -timeout 1s revisar $SRV/ok'

seccion "2.3 · Un subcomando mal escrito"
ejc 'vigia02 revsar $SRV/ok'

seccion "2.4 · La ayuda de un subcomando"
ejc 'vigia02 revisar -h'

seccion "2.5 · Cuánto código es el despachador"
ej 'grep -c "" 02_subcomandos/main.go'
