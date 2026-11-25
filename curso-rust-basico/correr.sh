#!/usr/bin/env bash
# Corre DENTRO del contenedor (lo lanza ejecutar_todo.py).
#   bash correr.sh                 todas las lecciones de todas las partes
#   bash correr.sh parte1/05       solo esa
set -u
rm -rf /work/repo; mkdir -p /work; cp -r /src /work/repo; cd /work/repo || exit 1
echo "@@VERSIONES"
rustc --version
cargo --version
uname -srm
if [ $# -eq 0 ]; then set -- $(ls -d parte*/[0-9][0-9]_*); fi
for pat in "$@"; do
	case "$pat" in */*) glob="${pat}*" ;; *) glob="${pat}/[0-9][0-9]_*" ;; esac
	for dir in $(ls -d $glob 2>/dev/null); do
		echo "@@INICIO $dir"
		bash "$dir/demo.sh" 2>&1
		echo "@@FIN $dir $?"
	done
done
