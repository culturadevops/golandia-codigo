# SALIDA — ¿SQL dentro de Go o en archivos .sql? (golandia 684)

Generado por `python correr.py` el 2026-10-27 12:48 UTC. **No se edita a mano**: cada bloque es el contenido de `salidas/<n>.txt`.

Postgres 18.6 (postgres:18.6-alpine) · Go 1.27.2 (golang:1.27.2) · sqlc 1.31.1 (sqlc/sqlc:1.31.1) · goose v3.29.0 · pgx v5.11.0 por database/sql · Docker 29.6.2. Todo en una red de Docker `g684`; el host de la base es `g684-pg`.

## 0 · Preparar: dos bases (app y tests) y dos usuarios (migrador y app)

```
$ psql -U postgres -f sql/roles.sql
CREATE ROLE
CREATE ROLE
CREATE DATABASE
CREATE DATABASE
You are now connected to database "tienda" as user "postgres".
GRANT
ALTER DEFAULT PRIVILEGES
You are now connected to database "tienda_test" as user "postgres".
GRANT
ALTER DEFAULT PRIVILEGES
[código de salida 0]
```

## 1 · Versiones exactas

```
$ docker version --format '{{.Server.Version}}'
29.6.2
[código de salida 0]
$ SELECT version()
                                         version                                         
-----------------------------------------------------------------------------------------
 PostgreSQL 18.6 on x86_64-pc-linux-musl, compiled by gcc (Alpine 15.2.0) 15.2.0, 64-bit
(1 row)
[código de salida 0]
$ go version
go version go1.27.2 linux/amd64
[código de salida 0]
$ goose --version
goose version: v3.29.0
[código de salida 0]
$ sqlc version
v1.31.1
[código de salida 0]
$ go list -m all | grep -E 'pgx|goose'
github.com/jackc/pgx/v5 v5.11.0
github.com/pressly/goose/v3 v3.29.0
[código de salida 0]
```

## 2 · ¿Una base separada? Una para la app, otra para los tests

```
$ \l  (las bases)
    base     |  duenio  
-------------+----------
 tienda      | migrador
 tienda_test | migrador
(2 rows)
[código de salida 0]
```

## 3 · Migraciones en archivos .sql con goose: status, up, tabla de versiones, down

```
$ goose -dir migrations postgres "$MIG" status
2026/10/27 12:47:55     Applied At                  Migration
2026/10/27 12:47:55     =======================================
2026/10/27 12:47:55     Pending                  -- 0001_crear_productos.sql
2026/10/27 12:47:55     Pending                  -- 0002_indice_categoria.sql
[código de salida 0]
$ goose -dir migrations postgres "$MIG" up
2026/10/27 12:47:55 OK   0001_crear_productos.sql (4.62ms)
2026/10/27 12:47:55 OK   0002_indice_categoria.sql (1.67ms)
2026/10/27 12:47:55 goose: successfully migrated database to version: 2
[código de salida 0]
$ goose -dir migrations postgres "$MIG" status
2026/10/27 12:47:56     Applied At                  Migration
2026/10/27 12:47:56     =======================================
2026/10/27 12:47:56     Tue Oct 27 12:47:55 2026 -- 0001_crear_productos.sql
2026/10/27 12:47:56     Tue Oct 27 12:47:55 2026 -- 0002_indice_categoria.sql
[código de salida 0]
$ SELECT * FROM goose_db_version
 id | version_id | is_applied 
----+------------+------------
  1 |          0 | t
  2 |          1 | t
  3 |          2 | t
(3 rows)
[código de salida 0]
$ goose -dir migrations postgres "$MIG" down
2026/10/27 12:47:56 OK   0002_indice_categoria.sql (2.59ms)
[código de salida 0]
$ goose -dir migrations postgres "$MIG" status
2026/10/27 12:47:56     Applied At                  Migration
2026/10/27 12:47:56     =======================================
2026/10/27 12:47:56     Tue Oct 27 12:47:55 2026 -- 0001_crear_productos.sql
2026/10/27 12:47:56     Pending                  -- 0002_indice_categoria.sql
[código de salida 0]
$ \d productos (tras el down: sigue la tabla, sin el índice)
                            Table "public.productos"
     Column      |  Type   | Collation | Nullable |           Default            
-----------------+---------+-----------+----------+------------------------------
 id              | bigint  |           | not null | generated always as identity
 nombre          | text    |           | not null | 
 categoria       | text    |           | not null | 
 precio_centavos | integer |           | not null | 
Indexes:
    "productos_pkey" PRIMARY KEY, btree (id)
Check constraints:
    "productos_precio_centavos_check" CHECK (precio_centavos >= 0)
[código de salida 0]
$ goose -dir migrations postgres "$MIG" up
2026/10/27 12:47:57 OK   0002_indice_categoria.sql (2.75ms)
2026/10/27 12:47:57 goose: successfully migrated database to version: 2
[código de salida 0]
$ SELECT * FROM goose_db_version
 id | version_id | is_applied 
----+------------+------------
  1 |          0 | t
  2 |          1 | t
  4 |          2 | t
(3 rows)
[código de salida 0]
```

## 4 · El usuario de la app no puede tocar el esquema

```
$ psql -U app -d tienda -f sql/semilla.sql
INSERT 0 8
[código de salida 0]
$ psql -U app: SELECT count(*) FROM productos
 count 
-------
     8
(1 row)
[código de salida 0]
$ psql -U app: DROP TABLE productos
ERROR:  must be owner of table productos
[código de salida 1]
$ psql -U app: CREATE TABLE copia (id int)
ERROR:  permission denied for schema public
LINE 1: CREATE TABLE copia (id int);
                     ^
[código de salida 1]
$ psql -U app: TRUNCATE productos
ERROR:  permission denied for table productos
[código de salida 1]
$ psql -U app: SELECT count(*) FROM productos
 count 
-------
     8
(1 row)
[código de salida 0]
```

## 5 · La misma consulta de tres formas: string, go:embed y sqlc

```
$ DATABASE_URL=postgres://app:…@g684-pg/tienda go run ./cmd/tres
conectado: ping OK
== (a) string en Go
   Café molido 250 g        7400
   Café soluble 200 g       9900
   Café molido 500 g       12900
== (b) .sql con go:embed
   Café molido 250 g        7400
   Café soluble 200 g       9900
   Café molido 500 g       12900
== (c) sqlc
   Café molido 250 g        7400
   Café soluble 200 g       9900
   Café molido 500 g       12900
[código de salida 0]
$ go build -trimpath -ldflags='-s -w' -o /tmp/tres ./cmd/tres && stat -c '%s bytes' /tmp/tres
10027271 bytes
[código de salida 0]
```

## 6 · Cambio una columna (nombre → titulo): quién avisa y cuándo

```
# copia de app/ en un directorio temporal; migración nueva 0003: nombre → titulo
-- +goose Up
ALTER TABLE productos RENAME COLUMN nombre TO titulo;

-- +goose Down
ALTER TABLE productos RENAME COLUMN titulo TO nombre;
$ goose -dir migrations postgres "$MIG" up
2026/10/27 12:47:59 OK   0001_crear_productos.sql (3.06ms)
2026/10/27 12:47:59 OK   0002_indice_categoria.sql (1.56ms)
2026/10/27 12:47:59 OK   0003_renombrar_nombre.sql (944.72µs)
2026/10/27 12:47:59 goose: successfully migrated database to version: 3
[código de salida 0]

# 1) sin regenerar nada: el código viejo COMPILA y falla al ejecutarse
$ go build ./... && go run ./cmd/tres
compila: OK
conectado: ping OK
== (a) string en Go
   ERROR: ERROR: column "nombre" does not exist (SQLSTATE 42703)
== (b) .sql con go:embed
   ERROR: ERROR: column "nombre" does not exist (SQLSTATE 42703)
== (c) sqlc
   ERROR: ERROR: column "nombre" does not exist (SQLSTATE 42703)
[código de salida 0]

# 1b) el test de integración, contra una base de tests con las migraciones nuevas
$ TEST_DATABASE_URL=…/tienda_cambio_test go test -count=1 ./consultas/
--- FAIL: TestLasTresFormasDanLoMismo (0.02s)
    consultas_test.go:53: ERROR: column "nombre" of relation "productos" does not exist (SQLSTATE 42703)
FAIL
FAIL	tienda/consultas	0.027s
FAIL
[código de salida 1]

# 2) sqlc generate con la consulta vieja
$ sqlc generate
# package db
query.sql:2:12: column "nombre" does not exist
[código de salida 1]

# 3) arreglo query.sql (nombre → titulo) y regenero
$ sqlc generate
[código de salida 0]
$ go build ./...
# tienda/cmd/tres
cmd/tres/main.go:32:93: c[i].Nombre undefined (type db.PorCategoriaRow has no field or method Nombre)
[código de salida 1]
```

## 7 · Consulta dinámica: estructura en Go, valores como $n; y la inyección

```
$ go run ./cmd/filtros
== sin filtros
   sql:  SELECT nombre FROM productos ORDER BY id
   args: []
   filas: 8
== categoria
   sql:  SELECT nombre FROM productos WHERE categoria = $1 ORDER BY id
   args: [cafe]
   filas: 5
== categoria + precio + texto
   sql:  SELECT nombre FROM productos WHERE categoria = $1 AND precio_centavos <= $2 AND nombre ILIKE $3 ORDER BY id
   args: [cafe 15000 %molido%]
   filas: 2

== entrada del usuario: ' OR '1'='1
-- concatenado
   sql:   SELECT nombre FROM productos WHERE categoria = '' OR '1'='1' ORDER BY id
   filas: 8  err: <nil>
-- placeholder
   sql:   SELECT nombre FROM productos WHERE categoria = $1 ORDER BY id
   args:  [' OR '1'='1]
   filas: 0  err: <nil>
[código de salida 0]
```

## 8 · La trampa de rows.Close() con un pool de 4 conexiones

```
$ ./pool olvido   (go build ./cmd/pool)
modo olvido · pool de 4 conexiones
  consulta 1  0.002s  en uso=1  Café molido 500 g
  consulta 2  0.007s  en uso=2  Café molido 500 g
  consulta 3  0.004s  en uso=3  Café molido 500 g
  consulta 4  0.004s  en uso=4  Café molido 500 g
  consulta 5  sigue esperando a los 5 s  en uso=4  esperando=1
[código de salida 3]
(reloj del contenedor: 5.4 s)
$ ./pool todo   (go build ./cmd/pool)
modo todo · pool de 4 conexiones
  consulta 1  0.002s  en uso=0  Café molido 500 g
  consulta 2  0.000s  en uso=0  Café molido 500 g
  consulta 3  0.000s  en uso=0  Café molido 500 g
  consulta 4  0.000s  en uso=0  Café molido 500 g
  consulta 5  0.000s  en uso=0  Café molido 500 g
  consulta 6  0.000s  en uso=0  Café molido 500 g
[código de salida 0]
(reloj del contenedor: 0.4 s)
$ ./pool ctx   (go build ./cmd/pool)
modo ctx · pool de 4 conexiones
  consulta 1  0.002s  en uso=1  Café molido 500 g
  consulta 2  0.005s  en uso=1  Café molido 500 g
  consulta 3  0.000s  en uso=1  Café molido 500 g
  consulta 4  0.000s  en uso=1  Café molido 500 g
  consulta 5  0.000s  en uso=1  Café molido 500 g
  consulta 6  0.000s  en uso=1  Café molido 500 g
[código de salida 0]
(reloj del contenedor: 0.4 s)
$ ./pool bien   (go build ./cmd/pool)
modo bien · pool de 4 conexiones
  consulta 1  0.002s  en uso=0  Café molido 500 g
  consulta 2  0.000s  en uso=0  Café molido 500 g
  consulta 3  0.000s  en uso=0  Café molido 500 g
  consulta 4  0.000s  en uso=0  Café molido 500 g
  consulta 5  0.000s  en uso=0  Café molido 500 g
  consulta 6  0.000s  en uso=0  Café molido 500 g
[código de salida 0]
(reloj del contenedor: 0.4 s)
$ SELECT count(*) FROM pg_stat_activity WHERE usename = 'app'
 count 
-------
     0
(1 row)
[código de salida 0]
```

## 9 · DSN por variable de entorno y Ping al arrancar

```
$ go run ./cmd/tres   (sin DATABASE_URL)
error al arrancar: falta DATABASE_URL
exit status 1
[código de salida 1]
$ DATABASE_URL=postgres://app:MAL@g684-pg/tienda go run ./cmd/tres
error al arrancar: no llegó a la base: failed to connect to `user=app database=tienda`: 192.168.192.2:5432 (g684-pg): failed SASL auth: FATAL: password authentication failed for user "app" (SQLSTATE 28P01)
exit status 1
[código de salida 1]
$ DATABASE_URL=postgres://app:app@g684-pg:5433/tienda go run ./cmd/tres   (puerto equivocado)
error al arrancar: no llegó a la base: failed to connect to `user=app database=tienda`: 192.168.192.2:5433 (g684-pg): dial error: dial tcp 192.168.192.2:5433: connect: connection refused
exit status 1
[código de salida 1]
```

## 10 · Test de integración contra la base de tests

```
$ go test -count=1 ./consultas/   (sin TEST_DATABASE_URL)
=== RUN   TestLasTresFormasDanLoMismo
    consultas_test.go:53: sin TEST_DATABASE_URL
--- SKIP: TestLasTresFormasDanLoMismo (0.00s)
PASS
ok  	tienda/consultas	0.004s
[código de salida 0]
$ TEST_DATABASE_URL=postgres://migrador:…@g684-pg/tienda_test go test -count=1 -v ./consultas/
=== RUN   TestLasTresFormasDanLoMismo
--- PASS: TestLasTresFormasDanLoMismo (0.03s)
PASS
ok  	tienda/consultas	0.031s
[código de salida 0]
$ psql -U app -d tienda: SELECT count(*) FROM productos   (la base de la app, intacta)
 count 
-------
     8
(1 row)
[código de salida 0]
```
