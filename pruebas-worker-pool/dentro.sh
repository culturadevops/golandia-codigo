#!/bin/sh
set -e

echo "== go version =="
go version

cd /lab/proyecto
go vet ./... 2>&1 || true

echo
echo "== sin control vs worker pool, corrido 3 veces =="
go run .
echo "---"
go run .
echo "---"
go run .
