#!/bin/bash
# medir.sh <restic|rustic> — corre DENTRO de gl-gra-backup:1 con --cpus 2 --memory 1g.
# /base (solo lectura): la carpeta de partida y el archivo nuevo del incremental.
# /datos, /repo, /rest: volúmenes de trabajo, vacíos al empezar.
# Cada paso imprime «PASO <nombre> seg=<reloj> rss_kb=<máx> user=<s> sys=<s> rc=<código>».
set -u
H="$1"
export RESTIC_PASSWORD=clave-del-banco RESTIC_REPOSITORY=/repo RESTIC_CACHE_DIR=/cache/restic
export RUSTIC_PASSWORD=clave-del-banco RUSTIC_REPOSITORY=/repo RUSTIC_CACHE_DIR=/cache/rustic
export RUSTIC_NO_PROGRESS=true
mkdir -p /cache
rm -rf /datos/* /repo/* /rest/* /cache/*
cp -a /base/datos/. /datos/
sync

paso() {   # paso <nombre> <comando...>
    local n="$1"; shift
    /usr/bin/time -o /tmp/t -f '%e %M %U %S' "$@" > /tmp/out.$n 2>&1
    local rc=$?
    read -r e m u s < <(tail -n 1 /tmp/t)
    echo "PASO $n seg=$e rss_kb=$m user=$u sys=$s rc=$rc"
    if [ $rc -ne 0 ]; then echo "--- salida de $n (rc=$rc):"; tail -n 20 /tmp/out.$n; fi
}
repo() { echo "REPO $1 bytes=$(du -sb /repo | cut -f1) archivos=$(find /repo -type f | wc -l)"; }

if [ "$H" = restic ]; then
    paso init      restic init
    paso inicial   restic backup --host banco /datos
    repo inicial
    paso sin_cambios restic backup --host banco /datos
elif [ "$H" = rustic ]; then
    paso init      rustic init
    paso inicial   rustic backup --host banco /datos
    repo inicial
    paso sin_cambios rustic backup --host banco /datos
fi
grep -E -i 'added to the repository|processed|Added|files:|snapshot' /tmp/out.inicial | head -8 | sed 's/^/  inicial> /'

# el mismo cambio para los dos: +1 línea en 1 de cada 50 archivos, se borra 1 de cada 97, entra un archivo nuevo de 32 MiB
find /datos -type f | LC_ALL=C sort > /tmp/lista
awk 'NR%50==0' /tmp/lista | while read -r f; do echo "cambio del banco" >> "$f"; done
awk 'NR%97==0' /tmp/lista | while read -r f; do rm -f "$f"; done
cp /base/extra/nuevo.bin /datos/binarios/nuevo.bin
echo "CAMBIO modificados=$(awk 'NR%50==0 && NR%97!=0' /tmp/lista | wc -l) borrados=$(awk 'NR%97==0' /tmp/lista | wc -l) nuevo=32MiB"
sync

if [ "$H" = restic ]; then
    paso incremental restic backup --host banco /datos
    repo incremental
    paso restaurar restic restore latest --target /rest
    paso check     restic check
    paso check_datos restic check --read-data
    restic snapshots --compact > /tmp/snaps 2>&1; echo "SNAPSHOTS $(grep -c '^[0-9a-f]\{8\} ' /tmp/snaps)"
    restic cat config 2>/dev/null | tr -d '\n ' | sed 's/^/CONFIG /'; echo
else
    paso incremental rustic backup --host banco /datos
    repo incremental
    paso restaurar rustic restore latest /rest
    paso check     rustic check
    paso check_datos rustic check --read-data
    rustic snapshots > /tmp/snaps 2>&1; echo "SNAPSHOTS $(grep -o '[0-9]* snapshot' /tmp/snaps | head -1)"
    rustic cat config 2>/dev/null | tr -d '\n ' | sed 's/^/CONFIG /'; echo
fi
grep -E -i 'added to the repository|Added|processed|files:' /tmp/out.incremental | head -6 | sed 's/^/  incremental> /'

# ¿llegó al mismo estado? lo restaurado contra la carpeta tal como quedó
R=/rest/datos; [ -d "$R" ] || R=/rest
D=$(diff -rq /datos "$R" 2>&1 | wc -l)
echo "RESTAURADO archivos=$(find "$R" -type f | wc -l) bytes=$(du -sb --apparent-size "$R" | cut -f1) diferencias=$D"
echo "ORIGEN archivos=$(find /datos -type f | wc -l) bytes=$(du -sb --apparent-size /datos | cut -f1)"
diff -rq /datos "$R" 2>&1 | head -5 | sed 's/^/  diff> /'
