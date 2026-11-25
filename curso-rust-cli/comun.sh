# Lo cargan los demo.sh. Cada versión de «fuga» se compila en /tmp/p/fuga con el Cargo.toml y el
# Cargo.lock de base/ (versiones fijadas) y su main.rs; las dependencias se compilan una vez (target compartido).
#   proyecto <main.rs> [tests.rs]   crea el proyecto, compila en debug y deja «fuga» en el PATH
#   mostrar '<orden>'               imprime la orden, la ejecuta (stdout+stderr) y su código de salida
#   ver <archivo>                   muestra un archivo de la lección
export CARGO_TERM_COLOR=never
export CARGO_TARGET_DIR=/work/target
AQUI_LECCION=$(cd "$(dirname "${BASH_SOURCE[1]}")" && pwd)
proyecto() {
	rm -rf /tmp/p/fuga; mkdir -p /tmp/p/fuga/src
	cp /work/repo/base/Cargo.toml /work/repo/base/Cargo.lock /tmp/p/fuga/
	cp "$AQUI_LECCION/$1" /tmp/p/fuga/src/main.rs
	if [ -n "${2:-}" ]; then mkdir -p /tmp/p/fuga/tests; cp "$AQUI_LECCION/$2" /tmp/p/fuga/tests/cli.rs; fi
	(cd /tmp/p/fuga && cargo build -q --locked 2>&1)
	ln -sf "$CARGO_TARGET_DIR/debug/fuga" /usr/local/bin/fuga
}
mostrar() { echo "\$ $1"; (cd /work/repo && eval "$1" 2>&1); echo "[codigo de salida: $?]"; }
ver() { echo "--- $1"; cat "$AQUI_LECCION/$1"; echo "---"; }
