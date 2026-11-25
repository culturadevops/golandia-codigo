#!/bin/bash
# Se ejecuta DENTRO del contenedor (golang:1.27.2 o golang:1.26). No lo corras a mano: lo lanza ejecutar_todo.py.
#   bash /trabajo/correr.sh demos   -> las demos de lenguaje, librería y herramientas
#   bash /trabajo/correr.sh banco   -> los bancos de malloc y flate (mediana de 5 la calcula ejecutar_todo.py)
# Todo se copia a un directorio de mktemp; nada se escribe en /trabajo.
set -u
export GOTOOLCHAIN=local GOFLAGS=-buildvcs=false
W=$(mktemp -d)
cp -r /trabajo/demos/. "$W"/
cd "$W" || exit 1
MENOR=$(go env GOVERSION | sed -E 's/^go1\.([0-9]+).*/\1/')
for d in */; do (cd "$d" && go mod edit -go=1.$MENOR); done

corre() { echo "\$ $*" | sed "s#$W#…#g";"$@" 2>&1 | sed "s#$W#…#g"; echo "[código de salida: ${PIPESTATUS[0]}]"; }
seccion() { echo; echo "## $1 · $(go env GOVERSION)"; cd "$W/$1" || exit 1; }

if [ "${1:-demos}" = "demos" ]; then
  echo "go version: $(go version)"
  seccion 01-metodos-genericos;  corre go run .
  seccion 02-campos-promovidos;  corre go run .
  seccion 03-inferencia;         corre go run .
  seccion 04-uuid;               corre go run .
  seccion 05-cutlast;            corre go run .
  seccion 06-goroutineleak;      corre go run .
  if [ "$MENOR" = "26" ]; then
    echo "# Go 1.26 con el experimento:"; corre env GOEXPERIMENT=goroutineleakprofile go run .
  else
    echo "# Go 1.27 con el experimento viejo:"; corre env GOEXPERIMENT=goroutineleakprofile go run .
  fi
  seccion 08-vet-w-puntero;      corre go run .; corre go vet .
  seccion 09-asynctimerchan
  corre cat go.mod;              corre go run .
  go mod edit -godebug=asynctimerchan=1
  corre cat go.mod;              corre go run .
  go mod edit -dropgodebug=asynctimerchan -godebug=asynctimerchan=0
  corre cat go.mod;              corre go run .
  seccion 11-traceback-etiquetas
  go build -o "$W/tb" .
  corre cat go.mod;              corre env GOTRACEBACK=all "$W/tb"
  echo "# El mismo programa con GODEBUG=tracebacklabels=0:"
  corre env GOTRACEBACK=all GODEBUG=tracebacklabels=0 "$W/tb"
  if [ "$MENOR" != "26" ]; then
    echo; echo "## Go 1.27 con «go 1.26» en go.mod (¿dependen de esa línea?)"
    cd "$W/01-metodos-genericos" && go mod edit -go=1.26 && corre go run .
    cd "$W/11-traceback-etiquetas" && go mod edit -go=1.26 && go build -o "$W/tb126" . && corre env GOTRACEBACK=all "$W/tb126"
  fi
  seccion 10-flate;              corre go test -run TestTamano -v .
  seccion 07-malloc
  cd hola
  go build -o "$W/hola-def" . && echo "binario hola (por defecto): $(stat -c %s "$W/hola-def") bytes"
  if [ "$MENOR" != "26" ]; then
    GOEXPERIMENT=nosizespecializedmalloc go build -o "$W/hola-no" . && \
      echo "binario hola (GOEXPERIMENT=nosizespecializedmalloc): $(stat -c %s "$W/hola-no") bytes"
  fi
else
  echo "go version: $(go version) · nproc=$(nproc)"
  seccion 07-malloc;  corre go test -run '^$' -bench Alloc -benchmem -count 5 .
  if [ "$MENOR" != "26" ]; then
    echo "# Go 1.27 sin las rutinas nuevas:"
    corre env GOEXPERIMENT=nosizespecializedmalloc go test -run '^$' -bench Alloc -benchmem -count 5 .
  fi
  seccion 10-flate;   corre go test -run '^$' -bench Gzip -count 5 .
fi
