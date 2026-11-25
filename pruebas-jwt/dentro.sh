#!/bin/sh
set -e

echo "== go version =="
go version

cd /lab/proyecto
go get github.com/golang-jwt/jwt/v5@v5.3.1 >/tmp/get1.log 2>&1 || (cat /tmp/get1.log; exit 1)
go get golang.org/x/crypto@latest >/tmp/get2.log 2>&1 || (cat /tmp/get2.log; exit 1)
go mod tidy >/tmp/tidy.log 2>&1 || (cat /tmp/tidy.log; exit 1)

echo
echo "== corrido =="
go run .
