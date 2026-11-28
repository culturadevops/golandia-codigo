#!/bin/bash
# compat.sh — ¿leen y escriben el MISMO formato de repositorio? Corre dentro de gl-gra-backup:1.
# A: repo creado por restic → rustic lo lee, lo comprueba, restaura y le añade un snapshot → restic lo comprueba.
# B: repo creado por rustic → restic lo lee, lo comprueba, restaura y le añade un snapshot → rustic lo comprueba.
set -u
export RESTIC_PASSWORD=clave-del-banco RUSTIC_PASSWORD=clave-del-banco RUSTIC_NO_PROGRESS=true
rm -rf /repo/* /rest/* /tmp/w; mkdir -p /tmp/w
cp -a /base/datos/go/src/net /tmp/w/net          # una carpeta chica de la misma base
N=$(find /tmp/w -type f | wc -l)
r() { local q="$1"; shift; "$@" > /tmp/o 2>&1; local rc=$?; echo "$q rc=$rc"; [ $rc -ne 0 ] && tail -n 8 /tmp/o | sed 's/^/   | /'; return 0; }

echo "## A · repo de restic"
mkdir -p /repo/a
export RESTIC_REPOSITORY=/repo/a RUSTIC_REPOSITORY=/repo/a
r "restic init" restic init
r "restic backup" restic backup --host banco /tmp/w
r "rustic snapshots" rustic snapshots; echo "   rustic ve: $(grep -o '[0-9]* snapshot[s]*' /tmp/o | head -1)"
r "rustic check --read-data" rustic check --read-data
rm -rf /rest/a; r "rustic restore" rustic restore latest /rest/a
echo "   diferencias tras restaurar con rustic: $(diff -rq /tmp/w /rest/a/tmp/w 2>&1 | wc -l) (archivos $(find /rest/a -type f | wc -l) de $N)"
echo "cambio" >> /tmp/w/net/http/server.go
r "rustic backup (en repo de restic)" rustic backup --host banco /tmp/w
r "restic snapshots" restic snapshots --compact; echo "   restic ve: $(grep -c '^[0-9a-f]\{8\} ' /tmp/o) snapshots"
r "restic check --read-data" restic check --read-data
rm -rf /rest/a2; r "restic restore (snapshot de rustic)" restic restore latest --target /rest/a2
echo "   diferencias tras restaurar con restic: $(diff -rq /tmp/w /rest/a2/tmp/w 2>&1 | wc -l)"

echo "## B · repo de rustic"
mkdir -p /repo/b
export RESTIC_REPOSITORY=/repo/b RUSTIC_REPOSITORY=/repo/b
r "rustic init" rustic init
r "rustic backup" rustic backup --host banco /tmp/w
r "restic snapshots" restic snapshots --compact; echo "   restic ve: $(grep -c '^[0-9a-f]\{8\} ' /tmp/o) snapshots"
r "restic check --read-data" restic check --read-data
rm -rf /rest/b; r "restic restore" restic restore latest --target /rest/b
echo "   diferencias tras restaurar con restic: $(diff -rq /tmp/w /rest/b/tmp/w 2>&1 | wc -l)"
echo "otro cambio" >> /tmp/w/net/http/server.go
r "restic backup (en repo de rustic)" restic backup --host banco /tmp/w
r "rustic snapshots" rustic snapshots; echo "   rustic ve: $(grep -o '[0-9]* snapshot[s]*' /tmp/o | head -1)"
r "rustic check --read-data" rustic check --read-data
echo "## config de cada repo"
echo "   restic init → $(RESTIC_REPOSITORY=/repo/a restic cat config 2>/dev/null | tr -d ' \n')"
echo "   rustic init → $(RESTIC_REPOSITORY=/repo/b restic cat config 2>/dev/null | tr -d ' \n')"
