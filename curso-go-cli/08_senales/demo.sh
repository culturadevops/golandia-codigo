#!/usr/bin/env bash
# Lección 8 — señales: Ctrl+C (SIGINT) y SIGTERM con context, sin dejar nada a medias.
source "$(dirname "$0")/../comun.sh"
compilar vigia ./vigia
compilar ingenuo08 ./08_senales/ingenuo
TMP="$(mktemp -d)"
cd "$TMP" || exit 1
LENTO="$SRV/lento?ms=5000"

# interrumpir SEÑAL comando...: lo lanza, a los 0,5 s le manda la señal (como Ctrl+C)
# y mide cuánto tarda en salir. Con set -m el proceso no hereda SIGINT ignorado, como
# pasa con los procesos en segundo plano de un script.
interrumpir() {
	local sen="$1"
	shift
	echo "\$ $*      # y a los 0,5 s: kill -$sen"
	set -m
	"$@" >/dev/null 2>"$TMP/stderr" &
	local pid=$!
	set +m
	sleep 0.5
	local t0 t1 c
	t0=$(date +%s%N)
	kill -"$sen" "$pid"
	wait "$pid"
	c=$?
	t1=$(date +%s%N)
	sed 's/^/  stderr: /' "$TMP/stderr"
	echo "  salió con código $c, $(( (t1 - t0) / 1000000 )) ms después de la señal"
}

estado() {
	echo "  informe.txt: $(wc -c < informe.txt) bytes -> $(head -c 60 informe.txt | tr '\n' '|')"
	echo "  temporales que quedaron: $(find . -name '.vigia-*' | wc -l)"
}

ayer() { printf 'informe de ayer: 2 de 2 arriba\n' > informe.txt; }

seccion "8.1 · Ingenuo, sin escuchar señales: Ctrl+C lo mata en el acto"
ayer
interrumpir INT ingenuo08 -informe informe.txt "$SRV/ok" "$LENTO"
estado

seccion "8.2 · Ingenuo con NotifyContext, pero la petición no lleva el ctx"
ayer
interrumpir INT ingenuo08 -escucha -informe informe.txt "$SRV/ok" "$LENTO"
estado

seccion "8.3 · vigia: ctx en cada petición e informe atómico"
ayer
interrumpir INT vigia revisar -informe informe.txt "$SRV/ok" "$LENTO"
estado

seccion "8.4 · vigia con SIGTERM (lo que manda docker stop o Kubernetes)"
ayer
interrumpir TERM vigia revisar -informe informe.txt "$SRV/ok" "$LENTO"
estado

seccion "8.5 · Y sin interrumpir, el informe nuevo sustituye al viejo de una vez"
ayer
ejc 'vigia revisar -informe informe.txt $SRV/ok $SRV/roto 2>/dev/null >/dev/null'
estado

cd / && rm -rf "$TMP"
