#!/usr/bin/env bash
# Parte 2: Postgres con pgx, migraciones, contexto con tiempo límite y pool de conexiones.
# La lanza ejecutar_todo.py con un Postgres (gl-curso-pg) en la misma red y DATABASE_URL puesta.
source /src/comun.sh
preparar parte2

echo "@@PASO dependencias"
x 'grep pgx go.mod'
x 'ls migraciones'
x 'go vet ./...'

echo "@@PASO sin-bd"
go build -o tareas . || exit 1
x 'DATABASE_URL= ./tareas'
x "DATABASE_URL='postgres://tareas:tareas@no-existe:5432/tareas' ./tareas"

echo "@@PASO primer-arranque"
arrancar
log_api
x "curl -s -X POST localhost:8080/tareas -d '{\"titulo\":\"comprar pan\"}'"
x "curl -s -X POST localhost:8080/tareas -d '{\"titulo\":\"llamar al dentista\"}'"
x "curl -s -X PATCH localhost:8080/tareas/1 -d '{\"hecha\":true}'"
x "curl -s -w 'HTTP %{http_code}\n' localhost:8080/tareas/7"

echo "@@PASO reinicio"
parar
arrancar
log_api
x "curl -s localhost:8080/tareas"

echo "@@PASO pool"
x "curl -s localhost:8080/salud"
x "seq 200 | xargs -P 50 -I{} curl -s -o /dev/null -w '%{http_code}\n' localhost:8080/tareas | sort | uniq -c"
x "curl -s localhost:8080/salud"

echo "@@PASO lenta"
go build -o bloquear ./herramientas/bloquear || exit 1
echo '$ ./bloquear 4 &'
./bloquear 4 > /tmp/bloquear.log 2>&1 &
BLOQ=$!
sleep 1.5
x "curl -s -w 'HTTP %{http_code} en %{time_total} s\n' localhost:8080/tareas"
x "curl -s -w 'HTTP %{http_code} en %{time_total} s\n' localhost:8080/salud"
wait $BLOQ
cat /tmp/bloquear.log
x "curl -s -w 'HTTP %{http_code} en %{time_total} s\n' localhost:8080/tareas"
log_api
parar
