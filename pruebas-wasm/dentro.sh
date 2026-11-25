#!/bin/sh
set -e

echo "== go version =="
go version
echo "== node version (mismo motor que usa Chrome) =="
node --version

cd /lab/proyecto
cp "$(go env GOROOT)/lib/wasm/wasm_exec.js" .

echo
echo "== 1. compilar Go a WebAssembly =="
GOOS=js GOARCH=wasm go build -o main.wasm main.go
ls -la main.wasm

echo
echo "== 2. el mismo codigo, compilado normal para linux (esto no deberia andar) =="
go build -o main-nativo main.go 2>&1 || echo "(fallo esperado: syscall/js no existe fuera de js/wasm)"

echo
echo "== 3. correr el .wasm en un motor de JavaScript, sin ningun servidor =="
node driver.js

echo
echo "== 4. lo mismo, pero con el binario reducido (sin simbolos de depuracion) =="
GOOS=js GOARCH=wasm go build -ldflags="-s -w" -o main-reducido.wasm main.go
ls -la main-reducido.wasm

echo
echo "== 5. lo que de verdad viaja por la red: comprimido =="
gzip -k -c main.wasm > main.wasm.gz
gzip -k -c main-reducido.wasm > main-reducido.wasm.gz
ls -la main.wasm.gz main-reducido.wasm.gz
