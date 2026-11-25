#!/usr/bin/env bash
# Corre DENTRO del contenedor (lo lanza ejecutar_todo.py).
#   bash correr.sh                 todas las lecciones de dentro del contenedor
#   bash correr.sh parte1/02       solo esa
set -u
rm -rf /work/repo; mkdir -p /work; cp -r /src /work/repo; cd /work/repo || exit 1
export CARGO_TARGET_DIR=/work/target CARGO_TERM_COLOR=never
echo "@@VERSIONES"
rustc --version
cargo --version
cargo build -q --release --locked 2>&1 | tail -5
grep -E '^name = "(tokio|reqwest|axum|serde_yaml|toml|tracing-subscriber)"' -A1 Cargo.lock | grep -v -- -- | paste - - | sed 's/name = //; s/version = //; s/"//g'
if [ $# -eq 0 ]; then set -- parte1 parte2 parte3; fi
for pat in "$@"; do
	case "$pat" in */*) glob="${pat}*" ;; *) glob="${pat}/[0-9][0-9]_*" ;; esac
	for dir in $(ls -d $glob 2>/dev/null); do
		[ -f "$dir/demo.sh" ] || continue
		echo "@@INICIO $dir"
		bash "$dir/demo.sh" 2>&1
		echo "@@FIN $dir $?"
	done
done
