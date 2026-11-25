# Lo cargan los demo.sh. Los binarios ya están compilados por correr.sh en /work/target/release.
#   ej <binario> [args...]   muestra la orden, la ejecuta y el código de salida
#   ver <archivo>            muestra un archivo de la lección o de src/bin
#   objetivo_arriba          arranca el servidor de pruebas en 127.0.0.1:8080
export PATH="/work/target/release:$PATH"
AQUI_LECCION=$(cd "$(dirname "${BASH_SOURCE[1]}")" && pwd)
cd "$AQUI_LECCION"
ej() {
	echo "\$ $*"
	"$@" 2>&1
	echo "[codigo de salida: $?]"
}
ver() { echo "--- $1"; cat "$1"; echo "---"; }
esperar_puerto() { for _ in $(seq 1 100); do curl -s -o /dev/null "http://127.0.0.1:$1/" && return 0; sleep 0.05; done; }
objetivo_arriba() { objetivo & OBJ=$!; esperar_puerto 8080; }
objetivo_abajo() { kill "$OBJ" 2>/dev/null; wait "$OBJ" 2>/dev/null; return 0; }
