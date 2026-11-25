#!/bin/sh
set -e

echo "== go version =="
go version

echo
echo "== instalar golangci-lint v2.13.2 =="
go install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v2.13.2 >/tmp/install.log 2>&1 || (cat /tmp/install.log; exit 1)
GOBIN="$(go env GOPATH)/bin"
"$GOBIN/golangci-lint" --version

echo
echo "== 1. golangci-lint contra el codigo real de kcli (el proyecto del video anterior) =="
cd /lab/repo
go mod tidy >/tmp/tidy.log 2>&1 || (cat /tmp/tidy.log; exit 1)
"$GOBIN/golangci-lint" run ./... || true

echo
echo "== 2. el mismo repo, con el error ya comprobado =="
cd /lab/repo-arreglado
go mod tidy >/tmp/tidy2.log 2>&1 || (cat /tmp/tidy2.log; exit 1)
"$GOBIN/golangci-lint" run ./... && echo "0 issues."

echo
echo "== 1b. si BindPFlag falla de verdad (nombre de bandera con un error de tipeo) =="
cd /lab/repo-falla-real
go mod tidy >/tmp/tidy4.log 2>&1 || (cat /tmp/tidy4.log; exit 1)
go run . serve start 2>&1 || true

echo
echo "== 2b. mas alla de errcheck: una funcion sin usar y una asignacion que se pisa =="
cd /lab/repo-mas-bugs
go mod tidy >/tmp/tidy3.log 2>&1 || (cat /tmp/tidy3.log; exit 1)
"$GOBIN/golangci-lint" run ./... || true

echo
echo "== 3. cuantos linters hay, y cuantos vienen activos por defecto =="
"$GOBIN/golangci-lint" linters > /tmp/all.txt 2>&1
echo "activos por defecto:"
sed -n '/^Enabled/,/^Disabled/p' /tmp/all.txt | grep -E "^[a-zA-Z0-9_]+:"
echo "total de linters disponibles: $(grep -cE '^[a-zA-Z0-9_]+:' /tmp/all.txt)"
