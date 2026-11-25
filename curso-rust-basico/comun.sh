# Lo cargan los demo.sh. Cada programa del curso se compila en un proyecto de Cargo de verdad,
# en /tmp/p/<nombre>, con el archivo .rs de la leccion como src/main.rs.
#   proyecto <nombre> <archivo.rs>      crea el proyecto
#   correr <nombre> [args...]           cargo run -q (debug) y el codigo de salida
#   correr_release <nombre> [args...]   cargo run -q --release
#   ver <archivo>                       muestra el codigo fuente
export CARGO_TERM_COLOR=never
AQUI_LECCION=$(cd "$(dirname "${BASH_SOURCE[1]}")" && pwd)
proyecto() {
	rm -rf "/tmp/p/$1"; mkdir -p /tmp/p
	(cd /tmp/p && cargo new -q --vcs none "$1")
	cp "$AQUI_LECCION/$2" "/tmp/p/$1/src/main.rs"
}
correr() {
	local n=$1; shift
	echo "\$ cargo run $*"
	(cd "/tmp/p/$n" && cargo run -q -- "$@" 2>&1)
	echo "[codigo de salida: $?]"
}
correr_release() {
	local n=$1; shift
	echo "\$ cargo run --release $*"
	(cd "/tmp/p/$n" && cargo run -q --release -- "$@" 2>&1)
	echo "[codigo de salida: $?]"
}
ver() { echo "--- $1"; cat "$AQUI_LECCION/$1"; echo "---"; }
