#!/bin/bash
# Dentro de golang:1.27.2: las cuatro estructuras del vídeo, compiladas y probadas de verdad.
cd /trabajo
apt-get -qq update >/dev/null 2>&1; apt-get -qq install -y tree >/dev/null 2>&1
echo '$ go version'; go version
echo; echo '# 1 · Plano: todo en la raíz'
cd /trabajo/plano
echo '$ tree'; tree --noreport
echo '$ go run .'; go run .
echo '$ go test ./...'; go test ./... 2>&1 | sed -E 's/\t[0-9.]+s$//'
echo; echo '# 2 · Servidor: cmd/ + internal/'
cd /trabajo/tienda
echo '$ tree'; tree --noreport
echo '$ go build ./... && go vet ./...'; go build ./... && go vet ./... && echo "(sin errores)"
echo '$ go test ./...'; go test ./... 2>&1 | sed -E 's/\t[0-9.]+s$//'
echo '$ go run ./cmd/api'; go run ./cmd/api
echo '$ go list ./...'; go list ./...
echo; echo '# 3 · Otro módulo importa tienda/internal/pedidos'
cd /trabajo/intruso
echo '$ go build .'; go build . 2>&1
echo; echo '# 3b · internal/ anidado: pedidos/internal/precio, importado desde cmd/api del MISMO módulo'
cd /trabajo/anidado
echo '$ tree'; tree --noreport
echo '$ go build ./pedidos/'; go build ./pedidos/ 2>&1 && echo "(sin errores)"
echo '$ go build ./cmd/api/'; go build ./cmd/api/ 2>&1
echo; echo '# 4 · models importa services y services importa models'
cd /trabajo/ciclo
echo '$ go build ./...'; go build ./... 2>&1
exit 0
