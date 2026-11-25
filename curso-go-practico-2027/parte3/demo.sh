#!/usr/bin/env bash
# Parte 3: pruebas de tabla, httptest, integración con Postgres en un contenedor, cobertura y -race.
# La lanza ejecutar_todo.py con un Postgres (gl-curso-pg) en la misma red; TAREAS_TEST_BD llega puesta
# como BD_PRUEBAS y aquí decidimos cuándo pasarla.
source /src/comun.sh
preparar parte3

echo "@@PASO tabla"
x 'go test -run TestValidarTitulo -v .'

echo "@@PASO rota"
sed -i 's/case n > maxTitulo:/case n >= maxTitulo:/' tareas.go
x 'grep -n "maxTitulo:" tareas.go'
x 'go test -run TestValidarTitulo .'
sed -i 's/case n >= maxTitulo:/case n > maxTitulo:/' tareas.go

echo "@@PASO httptest"
x 'go test -run "TestCrear|TestRutas|TestFlujo" -v .'

echo "@@PASO sin-bd"
x 'go test ./...'
x 'go test -run Postgres -v .'

echo "@@PASO con-bd"
x "TAREAS_TEST_BD=\"\$BD_PRUEBAS\" go test -run Postgres -v ."

echo "@@PASO cobertura"
x 'go test -cover .'
x "TAREAS_TEST_BD=\"\$BD_PRUEBAS\" go test -cover -coverprofile=cover.out ."
x 'go tool cover -func=cover.out'

echo "@@PASO carrera"
cp memoria.go /tmp/memoria.go
sed -i '/m.mu.Lock()/d; /m.mu.Unlock()/d' memoria.go
x 'grep -c "mu.Lock" memoria.go'
x 'go test -run TestMemoriaConcurrente -count=1 . 2>&1 | head -n 8'
x 'go test -race -run TestMemoriaConcurrente -count=1 . 2>&1 | head -n 30'
x 'go test -race -run TestMemoriaConcurrente -count=1 . 2>&1 | grep -E "^(FAIL|ok|--- )|race detected|DATA RACE" | sort | uniq -c'
cp /tmp/memoria.go memoria.go
x 'grep -c "mu.Lock" memoria.go'
x "TAREAS_TEST_BD=\"\$BD_PRUEBAS\" go test -race -count=1 ./..."
