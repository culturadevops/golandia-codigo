#!/usr/bin/env bash
# Parte 1: API REST con net/http, JSON, códigos de estado y validación. Almacén en memoria.
source /src/comun.sh
preparar parte1

echo "@@PASO versiones"
x 'go version'
x 'cat go.mod'
x 'ls'
x 'go vet ./...'

echo "@@PASO arrancar"
arrancar

echo "@@PASO crear"
x "curl -si -X POST localhost:8080/tareas -d '{\"titulo\":\"comprar pan\"}'"
x "curl -s -X POST localhost:8080/tareas -d '{\"titulo\":\"  llamar al dentista  \"}'"
x "curl -s localhost:8080/tareas"

echo "@@PASO validar"
x "curl -si -X POST localhost:8080/tareas -d '{\"titulo\":\"   \"}'"
x "curl -s -w 'HTTP %{http_code}\n' -X POST localhost:8080/tareas -d '{\"titulo\":\"'\$(printf 'a%.0s' \$(seq 121))'\"}'"
x "curl -s -w 'HTTP %{http_code}\n' -X POST localhost:8080/tareas -d '{\"titulo\":\"pan\"'"
x "curl -s -w 'HTTP %{http_code}\n' -X POST localhost:8080/tareas -d '{\"titel\":\"pan\"}'"
x "curl -s -w 'HTTP %{http_code}\n' -X POST localhost:8080/tareas -d '[1,2,3]'"

echo "@@PASO obtener"
x "curl -s -w 'HTTP %{http_code}\n' localhost:8080/tareas/1"
x "curl -s -w 'HTTP %{http_code}\n' localhost:8080/tareas/99"
x "curl -s -w 'HTTP %{http_code}\n' localhost:8080/tareas/abc"
x "curl -s -w 'HTTP %{http_code}\n' localhost:8080/tareas/-5"

echo "@@PASO marcar"
x "curl -s -w 'HTTP %{http_code}\n' -X PATCH localhost:8080/tareas/1 -d '{\"hecha\":true}'"
x "curl -s -w 'HTTP %{http_code}\n' -X PATCH localhost:8080/tareas/1 -d '{}'"
x "curl -s localhost:8080/tareas"

echo "@@PASO borrar"
x "curl -si -X DELETE localhost:8080/tareas/1"
x "curl -s -w 'HTTP %{http_code}\n' -X DELETE localhost:8080/tareas/1"

echo "@@PASO metodos"
x "curl -si -X PUT localhost:8080/tareas"
x "curl -si localhost:8080/"
x "curl -s -o /dev/null -w '%{http_code}\n' -I localhost:8080/tareas"

echo "@@PASO reiniciar"
parar
log_api
arrancar
x "curl -s localhost:8080/tareas"
parar
