#!/usr/bin/env bash
# Lección 10 — la versión dentro del binario: -ldflags -X y debug.ReadBuildInfo.
source "$(dirname "$0")/../comun.sh"
TMP="$(mktemp -d)"
cp -r . "$TMP/repo"
cd "$TMP/repo" || exit 1
rm -rf .git
git init -q
git -c user.name=curso -c user.email=curso@example.com add -A
git -c user.name=curso -c user.email=curso@example.com commit -qm "vigia: primera version"

seccion "10.1 · go run: sin datos del repositorio"
ej 'go run ./vigia version'

seccion "10.2 · go build dentro de un repositorio git: commit, fecha y si está limpio"
ej 'go build -o "$TMP/vigia" ./vigia && "$TMP/vigia" version'

seccion "10.3 · Con una etiqueta, go build pone la versión del módulo solo"
git -c user.name=curso -c user.email=curso@example.com tag v1.4.0
ej 'go build -o "$TMP/vigia" ./vigia && "$TMP/vigia" version | head -4'

seccion "10.4 · Un cambio sin commit: +dirty y vcs.modified"
echo "// cambio" >> vigia/run.go
ej 'go build -o "$TMP/vigia" ./vigia && "$TMP/vigia" version | grep -E "módulo|modified"'
git checkout -q vigia/run.go

seccion "10.5 · -ldflags -X: la versión que tú decides"
ej 'go build -ldflags "-X main.version=1.4.0 -X main.commit=$(git rev-parse --short HEAD) -X main.fecha=2026-10-28" -o "$TMP/vigia" ./vigia && "$TMP/vigia" version | head -2'

seccion "10.6 · Trampa: un nombre mal escrito en -X no da error"
ejc 'go build -ldflags "-X main.Version=1.4.0" -o "$TMP/vigia" ./vigia && "$TMP/vigia" version | head -1'
ejc 'go build -ldflags "-X curso-cli/vigia.version=1.4.0" -o "$TMP/vigia" ./vigia && "$TMP/vigia" version | head -1'

seccion "10.7 · Leerla desde fuera, sin ejecutar el binario"
ej 'go build -o "$TMP/vigia" ./vigia && go version -m "$TMP/vigia" | grep -E "path|mod|vcs\.(revision|modified)|GOOS|GOARCH"'

cd / && rm -rf "$TMP"
