#!/bin/bash
# Dentro de golang:1.27.2: compila y corre las demos de errgroup (golang.org/x/sync v0.24.0).
set -e
cd /trabajo/grupo && go mod tidy >/dev/null 2>&1 && go build -o /tmp/grupo .
echo '$ go version; grep x/sync go.mod'; go version; grep x/sync go.mod
for m in waitgroup errgroup sin-contexto dos-errores http http-sin-ctx pipeline pipeline-atascado limite; do
  echo; echo "\$ ./grupo $m"; /tmp/grupo $m
done
