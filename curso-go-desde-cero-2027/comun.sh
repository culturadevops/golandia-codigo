# Lo cargan los demo.sh. Cada programa del curso se compila en un módulo de Go de verdad,
# en /tmp/p/<nombre>, con el archivo .go de la lección como main.go.
#   proyecto <nombre> <archivo.go> [más archivos...]   crea el módulo
#   correr <nombre> [args...]                          go run . y el código de salida
#   ver <archivo>                                       muestra el código fuente
AQUI_LECCION=$(cd "$(dirname "${BASH_SOURCE[1]}")" && pwd)
export GOFLAGS=-buildvcs=false GOTOOLCHAIN=local
proyecto() {
	local n=$1 f=$2; shift 2
	rm -rf "/tmp/p/$n"; mkdir -p "/tmp/p/$n"
	cp "$AQUI_LECCION/$f" "/tmp/p/$n/main.go"
	for extra in "$@"; do cp "$AQUI_LECCION/$extra" "/tmp/p/$n/"; done
	(cd "/tmp/p/$n" && go mod init "$n" >/dev/null 2>&1)
}
correr() {
	local n=$1; shift
	echo "\$ go run . $*"
	(cd "/tmp/p/$n" && go run . "$@" 2>&1)
	echo "[codigo de salida: $?]"
}
ver() { echo "--- $1"; cat "$AQUI_LECCION/$1"; echo "---"; }
