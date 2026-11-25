#!/usr/bin/env bash
# Demo del vídeo 490 «Tu primera app de escritorio en Go con Fyne» — se EJECUTA y su salida va a SALIDA.md.
# En un contenedor (golang:1.27.2 + gcc + cabeceras de OpenGL/X11): crea el módulo, baja Fyne 2.8.2, compila la app
# (cgo) y corre una prueba que la usa sin pantalla y guarda capturas en capturas/.
set -u
export MSYS_NO_PATHCONV=1
D="$(cd "$(dirname "$0")" && pwd)"
mkdir -p "$D/capturas"
docker run --rm -v "$(cygpath -w "$D/app")":/src:ro -v "$(cygpath -w "$D/capturas")":/out -v golandia-gomod:/go/pkg/mod \
  golandia-fyne:2.8.2 sh -c '
  cp -r /src /app && cd /app
  go version
  echo "\$ go mod init mistareas && go get fyne.io/fyne/v2@v2.8.2"
  go mod init mistareas >/dev/null 2>&1 && go get fyne.io/fyne/v2@v2.8.2 2>&1 | grep -E "^go: added fyne" 
  echo "\$ go mod tidy   # sin esto, go build se queja de entradas que faltan en go.sum"
  go mod tidy >/dev/null 2>&1 && echo "go.sum: $(wc -l < go.sum) líneas (Fyne y sus dependencias)"
  echo "\$ time go build -o mistareas .   # primera vez: compila también la parte en C de Fyne"
  s=$(date +%s); go build -o mistareas . && echo "compilado en $(( $(date +%s) - s )) s"
  echo "\$ echo // cambio >> main.go && time go build -o mistareas ."
  echo "// cambio" >> main.go; s=$(date +%s%N); go build -o mistareas . && echo "recompilado en $(( ($(date +%s%N) - s) / 1000000 )) ms"
  echo "\$ ls -la mistareas"
  ls -la mistareas | awk "{print \$5, \$9}"
  echo "\$ go test -v -run TestUsarLaApp ."
  go test -v -run TestUsarLaApp . 2>&1 | grep -E "^(=== RUN|--- |    captura|ok|FAIL|PASS)|contador"
'
ls "$D/capturas"
