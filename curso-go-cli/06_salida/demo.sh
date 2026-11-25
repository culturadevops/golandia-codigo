#!/usr/bin/env bash
# Lección 6 — códigos de salida con significado, y cómo los lee Bash.
source "$(dirname "$0")/../comun.sh"
compilar vigia ./vigia
compilar ingenuo06 ./06_salida/ingenuo
TMP="$(mktemp -d)"

seccion "6.1 · Tres códigos, tres situaciones"
ejc 'vigia revisar $SRV/ok 2>/dev/null'
ejc 'vigia revisar $SRV/ok $SRV/roto 2>/dev/null'
ejc 'vigia revisar -timeout 3 $SRV/ok'

seccion "6.2 · Ingenuo: log.Fatal en la configuración da el mismo 1 que un caído"
ejc 'ingenuo06 $SRV/roto'
ejc 'ingenuo06 -timeout 3 $SRV/ok'

seccion "6.3 · Ingenuo: os.Exit se salta el defer, y el búfer se pierde"
ejc 'ingenuo06 $SRV/ok'
ejc 'ingenuo06 $SRV/ok $SRV/roto'
ej 'ingenuo06 $SRV/ok $SRV/roto | wc -l'

seccion "6.4 · En un script: if y case sobre el código"
cat > "$TMP/vigilar.sh" <<'FIN'
vigia revisar "$@" 2>/dev/null
case $? in
  0) echo "todo arriba" ;;
  1) echo "hay caídos: aviso a guardia" ;;
  2) echo "el script está mal llamado: esto lo arregla quien lo escribió" ;;
  *) echo "otra cosa: interrumpido o algo que no esperaba" ;;
esac
FIN
ej 'bash "$TMP/vigilar.sh" $SRV/ok'
ej 'bash "$TMP/vigilar.sh" $SRV/ok $SRV/roto'
ej 'bash "$TMP/vigilar.sh" -timeout 3 $SRV/ok'

seccion "6.5 · Trampa de Bash: en un pipe, el código es el del ÚLTIMO comando"
ejc 'vigia revisar $SRV/roto 2>/dev/null | tee "$TMP/informe.txt" > /dev/null'
ejc 'set -o pipefail; vigia revisar $SRV/roto 2>/dev/null | tee "$TMP/informe.txt" > /dev/null'
set +o pipefail
ej 'vigia revisar $SRV/roto 2>/dev/null | tee "$TMP/informe.txt" > /dev/null; echo "PIPESTATUS: ${PIPESTATUS[*]}"'

seccion "6.6 · set -e tampoco ve lo que pasa a la izquierda del pipe"
ej 'bash -c "set -e; vigia revisar $SRV/roto 2>/dev/null | wc -l; echo sigo vivo"'
ej 'bash -c "set -eo pipefail; vigia revisar $SRV/roto 2>/dev/null | wc -l; echo sigo vivo"; echo "[código de salida: $?]"'

rm -rf "$TMP"
