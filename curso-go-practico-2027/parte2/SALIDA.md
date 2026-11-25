# SALIDA — Go práctico 2027 · parte 2

> Generado por `ejecutar_todo.py`. **No se escribe a mano.** Ids, tiempos y tamaños son los de esta corrida.

- fecha: `2026-11-23 12:03:23Z`
- imagen Go: `golang:1.27.2` · Postgres: `postgres:18.6-alpine` · `Docker 29.6.2`

## dependencias

```
$ grep pgx go.mod
require github.com/jackc/pgx/v5 v5.13.0
$ ls migraciones
001_crear_tareas.sql
002_fecha_de_creacion.sql
$ go vet ./...
```

## sin-bd

```
$ DATABASE_URL= ./tareas
2026/11/23 12:03:09 falta DATABASE_URL
[código de salida: 1]
$ DATABASE_URL='postgres://tareas:tareas@no-existe:5432/tareas' ./tareas
2026/11/23 12:03:17 no llego a Postgres: context deadline exceeded
[código de salida: 1]
```

## primer-arranque

```
$ ./tareas &
$ cat api.log
2026/11/23 12:03:17 migraciones aplicadas: [001_crear_tareas.sql 002_fecha_de_creacion.sql]
2026/11/23 12:03:17 escuchando en :8080
$ curl -s -X POST localhost:8080/tareas -d '{"titulo":"comprar pan"}'
{"id":1,"titulo":"comprar pan","hecha":false}
$ curl -s -X POST localhost:8080/tareas -d '{"titulo":"llamar al dentista"}'
{"id":2,"titulo":"llamar al dentista","hecha":false}
$ curl -s -X PATCH localhost:8080/tareas/1 -d '{"hecha":true}'
{"id":1,"titulo":"comprar pan","hecha":true}
$ curl -s -w 'HTTP %{http_code}\n' localhost:8080/tareas/7
{"error":"tarea no encontrada"}
HTTP 404
```

## reinicio

```
$ kill -TERM $PID
[la API terminó con código 143]
$ ./tareas &
$ cat api.log
2026/11/23 12:03:17 migraciones aplicadas: []
2026/11/23 12:03:17 escuchando en :8080
$ curl -s localhost:8080/tareas
[{"id":1,"titulo":"comprar pan","hecha":true},{"id":2,"titulo":"llamar al dentista","hecha":false}]
```

## pool

```
$ curl -s localhost:8080/salud
{"bd":"ok","conexiones_abiertas":1,"en_uso":0,"max":4,"esperas":1}
$ seq 200 | xargs -P 50 -I{} curl -s -o /dev/null -w '%{http_code}\n' localhost:8080/tareas | sort | uniq -c
    200 200
$ curl -s localhost:8080/salud
{"bd":"ok","conexiones_abiertas":4,"en_uso":0,"max":4,"esperas":10}
```

## lenta

```
$ ./bloquear 4 &
$ curl -s -w 'HTTP %{http_code} en %{time_total} s\n' localhost:8080/tareas
{"error":"la base de datos no respondió a tiempo"}
HTTP 503 en 2.002102 s
$ curl -s -w 'HTTP %{http_code} en %{time_total} s\n' localhost:8080/salud
{"bd":"ok","conexiones_abiertas":3,"en_uso":0,"max":4,"esperas":10}
HTTP 200 en 0.001024 s
bloquear: tabla tareas bloqueada 4 s
bloquear: tabla tareas liberada
$ curl -s -w 'HTTP %{http_code} en %{time_total} s\n' localhost:8080/tareas
[{"id":1,"titulo":"comprar pan","hecha":true},{"id":2,"titulo":"llamar al dentista","hecha":false}]
HTTP 200 en 0.000792 s
$ cat api.log
2026/11/23 12:03:21 error del almacén: la base de datos no respondió a tiempo: timeout: context deadline exceeded
$ kill -TERM $PID
[la API terminó con código 143]

[demo.sh terminó con código 0]
```
