#!/usr/bin/env bash
# Corre DENTRO del contenedor (lo lanza ejecutar_todo.py, con --memory=1g --cpus=2 --shm-size=256m).
#   bash correr.sh                 todas las lecciones
#   bash correr.sh parte2/02       solo esa
set -u
rm -rf /work/repo; mkdir -p /work; cp -r /src /work/repo; cd /work/repo || exit 1
echo "@@VERSIONES"
rustc --version
cargo --version
uname -srm
echo "docker run --memory=1g --memory-swap=1g --cpus=2 --shm-size=256m"
(cd opstool && cargo build --release --locked -q 2>&1) || exit 1
export PATH="/work/repo/opstool/target/release:$PATH"
if [ $# -eq 0 ]; then set -- parte1 parte2 parte3; fi
for pat in "$@"; do
	case "$pat" in */*) glob="${pat}*" ;; *) glob="${pat}/[0-9][0-9]_*" ;; esac
	for dir in $(ls -d $glob 2>/dev/null); do
		echo "@@INICIO $dir"
		(cd "$dir" && bash ./demo.sh) 2>&1
		echo "@@FIN $dir $?"
	done
done
