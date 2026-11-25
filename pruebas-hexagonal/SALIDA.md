# SALIDA — Arquitectura hexagonal en Go (golandia)

> Corrida del 2026-10-19 16:26. Salida literal de correr.sh dentro de golang:1.26, con Postgres 17.

```
go version go1.26.9 linux/amd64
### 1 · Lo que escribes: la misma API dos veces
$ cd plana && find . -name '*.go' | xargs wc -l
70 ./main.go
$ cd hexagonal && find . -name '*.go' -not -name '*_test.go' | xargs wc -l
  29 ./cmd/api/main.go
  36 ./internal/adaptadores/memoria/memoria.go
  31 ./internal/adaptadores/postgres/postgres.go
  54 ./internal/adaptadores/web/web.go
  56 ./internal/pedidos/pedidos.go
 206 total

### 2 · Las dos responden igual (contra Postgres)
--- plana
$ curl -s -X POST localhost:8080/pedidos -d '{"cliente":"Ana","unidades":3,"precio":4500}'
{"id":1,"cliente":"Ana","unidades":3,"precio":4500,"total":13500}
$ curl -s -X POST localhost:8080/pedidos -d '{"cliente":"","unidades":3,"precio":4500}'
el pedido necesita un cliente
--- hex
$ curl -s -X POST localhost:8080/pedidos -d '{"cliente":"Ana","unidades":3,"precio":4500}'
{"id":2,"cliente":"Ana","unidades":3,"precio":4500,"total":13500}
$ curl -s -X POST localhost:8080/pedidos -d '{"cliente":"","unidades":3,"precio":4500}'
el pedido necesita un cliente

### 3 · Probar las reglas: 1.000 rondas de 4 casos, con el adaptador en memoria y con Postgres
$ go test -count=1 -run TestReglasEnMemoria -v ./internal/pedidos
--- PASS: TestReglasEnMemoria (0.00s)
ok  	tienda/internal/pedidos	0.003s
$ go test -count=1 -run TestReglasEnPostgres -v ./internal/pedidos
--- PASS: TestReglasEnPostgres (0.76s)
ok  	tienda/internal/pedidos	0.760s

### 4 · Cambiar de adaptador: sin DATABASE_URL, la misma app guarda en memoria
$ curl -s -X POST localhost:8080/pedidos -d '{"cliente":"Luis","unidades":2,"precio":1000}'
{"id":1,"cliente":"Luis","unidades":2,"precio":1000,"total":2000}
$ curl -s localhost:8080/pedidos/1
{"id":1,"cliente":"Luis","unidades":2,"precio":1000,"total":2000}
```
