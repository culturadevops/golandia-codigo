#!/bin/bash
# Corre DENTRO de golang:1.26 en la red «hexagonal», con Postgres 17 en el host «pg» (lo lanza ejecutar.sh).
set -u
export DATABASE_URL=postgres://tienda:tienda@pg:5432/tienda
rm -rf /w && mkdir -p /w && cp -r /trabajo/plana /trabajo/hexagonal /w/
for d in plana hexagonal; do
  (cd /w/$d && go mod init tienda >/dev/null 2>&1 && go get github.com/jackc/pgx/v5@v5.7.6 >/dev/null 2>&1 && go mod tidy >/dev/null 2>&1)
done

echo "### 1 · Lo que escribes: la misma API dos veces"
echo "\$ cd plana && find . -name '*.go' | xargs wc -l"
(cd /w/plana && find . -name '*.go' | sort | xargs wc -l)
echo "\$ cd hexagonal && find . -name '*.go' -not -name '*_test.go' | xargs wc -l"
(cd /w/hexagonal && find . -name '*.go' -not -name '*_test.go' | sort | xargs wc -l)

echo
echo "### 2 · Las dos responden igual (contra Postgres)"
(cd /w/plana && go build -o /tmp/plana . ) && (cd /w/hexagonal && go build -o /tmp/hex ./cmd/api) || exit 1
for app in plana hex; do
  PORT=8080 /tmp/$app > /tmp/$app.log 2>&1 &
  PID=$!
  for i in $(seq 1 50); do curl -s -o /dev/null localhost:8080/pedidos/0 && break; sleep 0.1; done
  echo "--- $app"
  echo "\$ curl -s -X POST localhost:8080/pedidos -d '{\"cliente\":\"Ana\",\"unidades\":3,\"precio\":4500}'"
  curl -s -X POST localhost:8080/pedidos -d '{"cliente":"Ana","unidades":3,"precio":4500}'
  echo "\$ curl -s -X POST localhost:8080/pedidos -d '{\"cliente\":\"\",\"unidades\":3,\"precio\":4500}'"
  curl -s -X POST localhost:8080/pedidos -d '{"cliente":"","unidades":3,"precio":4500}'
  kill $PID; wait $PID 2>/dev/null
done

echo
echo "### 3 · Probar las reglas: 1.000 rondas de 4 casos, con el adaptador en memoria y con Postgres"
cd /w/hexagonal
go test -count=1 -run XXX ./internal/pedidos >/dev/null 2>&1   # compila las pruebas antes de cronometrar
echo "\$ go test -count=1 -run TestReglasEnMemoria -v ./internal/pedidos"
go test -count=1 -run TestReglasEnMemoria -v ./internal/pedidos 2>&1 | grep -E "^(--- |ok|FAIL)"
echo "\$ go test -count=1 -run TestReglasEnPostgres -v ./internal/pedidos"
go test -count=1 -run TestReglasEnPostgres -v ./internal/pedidos 2>&1 | grep -E "^(--- |ok|FAIL)"

echo
echo "### 4 · Cambiar de adaptador: sin DATABASE_URL, la misma app guarda en memoria"
unset DATABASE_URL
/tmp/hex > /tmp/hexm.log 2>&1 &
PID=$!
for i in $(seq 1 50); do curl -s -o /dev/null localhost:8080/pedidos/0 && break; sleep 0.1; done
echo "\$ curl -s -X POST localhost:8080/pedidos -d '{\"cliente\":\"Luis\",\"unidades\":2,\"precio\":1000}'"
curl -s -X POST localhost:8080/pedidos -d '{"cliente":"Luis","unidades":2,"precio":1000}'
echo "\$ curl -s localhost:8080/pedidos/1"
curl -s localhost:8080/pedidos/1
kill $PID
