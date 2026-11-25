# Funciones que usan los demo.sh de cada lección. No se ejecuta solo: se carga con «source».
#
# Lo que hace:
#   · deja los binarios del curso en $BIN y los mete en el PATH;
#   · levanta el servidor de pruebas (./banco) en $SRV si no está ya;
#   · ej 'comando'   imprime el comando tal cual y lo ejecuta;
#   · ejc 'comando'  igual, y además imprime el código de salida.

set -u
RAIZ="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
cd "$RAIZ" || exit 1
export BIN="${BIN:-/tmp/curso-cli-bin}"
mkdir -p "$BIN"
export PATH="$BIN:$PATH"
export SRV="http://127.0.0.1:8080"

compilar() { # compilar NOMBRE PAQUETE [flags de go build...]
	local nombre="$1" paquete="$2"
	shift 2
	go build "$@" -o "$BIN/$nombre" "$paquete" || exit 1
}

seccion() { echo; echo "== $* =="; }

ej() { echo "\$ $1"; eval "$1"; }

ejc() {
	echo "\$ $1"
	eval "$1"
	echo "[código de salida: $?]"
}

if ! curl -fs "$SRV/ok" >/dev/null 2>&1; then
	compilar banco ./banco
	banco 2>/dev/null &
	for _ in $(seq 50); do
		curl -fs "$SRV/ok" >/dev/null 2>&1 && break
		sleep 0.1
	done
fi
