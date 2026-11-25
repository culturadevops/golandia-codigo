#!/bin/bash
# Demo del vídeo «Goravel: ¿el Laravel de Go?» (pedido en un comentario del vídeo de PHP, 2026-08-03).
# Corre DENTRO de golang:1.26, en la red «goravel» junto a un Postgres 17 (lo levanta correr.py).
# Todo lo que sale en el vídeo se copia de SALIDA.md, que es la salida literal de esto.
set -u
rm -rf /w && mkdir -p /w
cp -r /trabajo/esqueleto /w/app && cd /w/app
rm -rf .git .agents
cp .env.example .env
cat >> .env <<'EOF'
DB_HOST=pg
DB_PORT=5432
DB_DATABASE=tienda
DB_USERNAME=goravel
DB_PASSWORD=goravel
APP_HOST=0.0.0.0
EOF

echo "### 1 · Lo que trae el esqueleto (goravel/goravel v1.18.0)"
echo "\$ find . -name '*.go' | wc -l"
find . -name '*.go' | wc -l
echo "\$ go mod download && go list -m all | wc -l   # módulos en el árbol de dependencias"
go mod download 2>/dev/null
go list -m all | wc -l

echo
echo "### 2 · Compilar (caché vacía) y el binario"
echo "\$ time go build -o /tmp/app ."
{ time -p go build -o /tmp/app . ; } 2>&1 | grep real
echo "\$ ls -l /tmp/app"
ls -l /tmp/app | awk '{print $5, "bytes"}'

echo
echo "### 3 · artisan: los generadores de Laravel, en Go"
echo "\$ ./artisan key:generate"
go run . artisan key:generate 2>&1 | tail -2
echo "\$ ./artisan make:controller ProductoController"
go run . artisan make:controller ProductoController 2>&1 | tail -2
echo "\$ ./artisan make:model Producto"
go run . artisan make:model Producto 2>&1 | tail -2
echo "\$ ./artisan make:migration create_productos_table"
go run . artisan make:migration create_productos_table 2>&1 | tail -2
echo "\$ ls app/http/controllers app/models database/migrations"
ls app/http/controllers app/models database/migrations
echo "\$ cat app/models/producto.go"
cat app/models/producto.go
