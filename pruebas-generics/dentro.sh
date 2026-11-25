#!/bin/sh
set -e

echo "== go version =="
go version

echo
echo "== 1. benchmark: generico vs a mano vs interface{} =="
cd /lab/etapas/01-benchmark
go test -run=^$ -bench=. -benchmem -benchtime=200000x

echo
echo "== 1b. slices: la libreria estandar de Go, escrita con generics =="
cd /lab/etapas/03-stdlib
go run .

echo
echo "== 2. seguridad de tipos: la version buena, corre =="
cd /lab/etapas/02-seguridad-tipos/bien
go run .

echo
echo "== 3. seguridad de tipos: string en SumGeneric, NO compila =="
cd /lab/etapas/02-seguridad-tipos/mal
go build . 2>&1 || true

echo
echo "== 4. la virgulilla (~): con ella, un tipo nombrado tambien sirve =="
cd /lab/etapas/04-tilde/bien
go run .

echo
echo "== 5. sin la virgulilla: el mismo tipo nombrado, y NO compila =="
cd /lab/etapas/04-tilde/mal
go build . 2>&1 || true
