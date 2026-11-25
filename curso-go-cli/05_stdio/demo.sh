#!/usr/bin/env bash
# Lección 5 — stdin, stdout y stderr: los datos por un lado, lo demás por otro.
source "$(dirname "$0")/../comun.sh"
compilar vigia ./vigia
compilar ingenuo05 ./05_stdio/ingenuo
TMP="$(mktemp -d)"
unset NO_COLOR

seccion "5.1 · En la terminal se ve todo junto"
ej 'vigia revisar $SRV/ok $SRV/roto'

seccion "5.2 · Pero son dos canales: stderr fuera"
ej 'vigia revisar $SRV/ok $SRV/roto 2>/dev/null'

seccion "5.3 · Contar líneas en un pipe"
ej 'vigia revisar $SRV/ok $SRV/roto 2>/dev/null | wc -l'
ej 'ingenuo05 $SRV/ok $SRV/roto 2>/dev/null | wc -l'

seccion "5.4 · Guardar en un archivo: lo que queda dentro"
ej 'vigia revisar $SRV/ok 2>/dev/null > "$TMP/bueno.txt"; cat -v "$TMP/bueno.txt"'
ej 'ingenuo05 $SRV/ok > "$TMP/ingenuo.txt"; cat -v "$TMP/ingenuo.txt"'

seccion "5.5 · Con una terminal de verdad delante (script le da una), vigia sí pinta"
ej 'script -qc "vigia revisar $SRV/ok 2>/dev/null" /dev/null | cat -v'
ej 'NO_COLOR=1 script -qc "vigia revisar $SRV/ok 2>/dev/null" /dev/null | cat -v'

seccion "5.6 · Trampa: ¿es una terminal? /dev/null también es un dispositivo de caracteres"
compilar esterminal ./05_stdio/esterminal
ej 'esterminal | cat'
ej 'esterminal > "$TMP/archivo.txt"'
ej 'esterminal > /dev/null'
ej 'script -qc "esterminal" /dev/null'

seccion "5.7 · Las URLs por stdin, con «-»"
ej 'cat 05_stdio/urls.txt'
ejc 'vigia revisar - < 05_stdio/urls.txt'
ejc 'grep -v roto 05_stdio/urls.txt | vigia revisar - 2>/dev/null'

rm -rf "$TMP"
