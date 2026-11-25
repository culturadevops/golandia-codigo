#!/bin/sh
# Corre dentro de golang:1.27.2. /w = la carpeta del vídeo, montada de solo lectura.
go version
go install honnef.co/go/tools/cmd/staticcheck@v0.8.2 >/dev/null 2>&1
staticcheck -version
for d in go go121 go127 go_bien; do
  cd /w/casos/$d
  echo "=== caso|$d|vet"; go vet ./... 2>&1; echo "--- exit=$?"
  echo "=== caso|$d|staticcheck"; staticcheck ./... 2>&1; echo "--- exit=$?"
  echo "=== caso|$d|run"; go run . 2>&1; echo "--- exit=$?"
done
cd /w/casos/go121
echo "=== caso|go121|race"; go run -race . > /tmp/race.txt 2>&1; echo "--- exit=$?"
echo "líneas «WARNING: DATA RACE»: $(grep -c 'WARNING: DATA RACE' /tmp/race.txt)"; tail -3 /tmp/race.txt
cd /w/casos/go127
echo "=== caso|go127|fix"; go fix -diff ./... 2>&1; echo "--- exit=$?"
for f in /w/ia/*/go/*.go; do
  t=$(mktemp -d); cp "$f" "$t/main.go"; cd "$t"; go mod init muestra >/dev/null 2>&1
  rel=${f#/w/}
  echo "=== ia|$rel|vet"; go vet ./... 2>&1; echo "--- exit=$?"
  echo "=== ia|$rel|staticcheck"; staticcheck ./... 2>&1; echo "--- exit=$?"
done
