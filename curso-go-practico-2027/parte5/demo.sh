#!/usr/bin/env bash
# Parte 5 (dentro del contenedor): el fallo que encontró la BD caída, en una prueba.
# La misma prueba de clasificar contra el código de la parte 4 (falla) y el de la parte 5 (pasa).
# Lo demás (compose, health checks, la BD que se cae) lo hace ejecutar_todo.py desde fuera.
source /src/comun.sh

echo "@@PASO clasificar-parte4"
rm -rf /work/p4 && mkdir -p /work && cp -r /src/parte4/tareas /work/p4 && cp /src/parte5/tareas/clasificar_test.go /work/p4/
cd /work/p4 || exit 1
x 'go test -run TestClasificar .'

echo "@@PASO clasificar-parte5"
preparar parte5
x 'grep -n "conexionPerdida" -A2 postgres.go | head -n 4'
x 'go test -run TestClasificar -v .'
x 'go vet ./... && go test -race -count=1 ./...'
