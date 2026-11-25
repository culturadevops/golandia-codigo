#!/bin/sh
set -e

echo "== go version =="
go version

echo
echo "== instalar goreleaser v2.18.1 =="
go install github.com/goreleaser/goreleaser/v2@v2.18.1 >/tmp/install.log 2>&1 || (cat /tmp/install.log; exit 1)
GOBIN="$(go env GOPATH)/bin"
"$GOBIN/goreleaser" --version | head -8

cd /lab/proyecto
git init -q
git config user.email demo@example.com
git config user.name demo
go mod tidy >/tmp/tidy.log 2>&1 || (cat /tmp/tidy.log; exit 1)
git add -A
git commit -q -m "feat: primera version de kcli"
git commit -q --allow-empty -m "feat: agrega el comando serve"
git commit -q --allow-empty -m "fix: corrige el puerto por defecto"
git tag v0.1.0

echo
echo "== 1. build normal de Go: un binario, solo para esta maquina =="
go build -o /tmp/kcli-normal .
ls -la /tmp/kcli-normal

echo
echo "== 2. goreleaser en modo real (release --clean), SIN token ni remoto: falla, y dice por que =="
"$GOBIN/goreleaser" release --clean 2>&1 | tail -6 || true

echo
echo "== 3. goreleaser en modo snapshot: construye todo, no publica nada =="
"$GOBIN/goreleaser" release --snapshot --clean 2>&1 | tail -25

echo
echo "== 4. que produjo, archivo por archivo =="
find dist -maxdepth 1 -type f | sort

echo
echo "== 5. seis binarios reales, para tres sistemas operativos =="
find dist -name "kcli" -o -name "kcli.exe" | sort

echo
echo "== 6. el changelog: por que no aparece en modo snapshot =="
cat dist/CHANGELOG.md 2>/dev/null || echo "(no existe CHANGELOG.md: el pipeline de changelog se salta a proposito en snapshot)"
