# SALIDA — Go práctico 2027 · parte 5

> Generado por `ejecutar_todo.py`. **No se escribe a mano.** Ids, tiempos y tamaños son los de esta corrida.

- fecha: `2026-11-23 12:16:39Z`
- imagen Go: `golang:1.27.2` · Postgres: `postgres:18.6-alpine` · `Docker 29.6.2`

## clasificar-parte4

```
$ go test -run TestClasificar .
--- FAIL: TestClasificar (0.00s)
    --- FAIL: TestClasificar/Postgres_apagándose (0.00s)
        clasificar_test.go:32: clasificar(:  (SQLSTATE 57P01)) = :  (SQLSTATE 57P01), quiero base de datos no disponible
    --- FAIL: TestClasificar/conexión_rota (0.00s)
        clasificar_test.go:32: clasificar(:  (SQLSTATE 08006)) = :  (SQLSTATE 08006), quiero base de datos no disponible
FAIL
FAIL	tareas	0.002s
FAIL
[código de salida: 1]
```

## clasificar-parte5

```
$ grep -n "conexionPerdida" -A2 postgres.go | head -n 4
125:	case errors.As(err, &pgErr) && conexionPerdida(pgErr.Code):
126-		return fmt.Errorf("%w: %w", ErrBDCaida, err)
127-	case errors.As(err, &pgErr):
--
$ go test -run TestClasificar -v .
=== RUN   TestClasificar
=== RUN   TestClasificar/sin_filas
=== RUN   TestClasificar/tiempo_agotado
=== RUN   TestClasificar/check_de_la_tabla
=== RUN   TestClasificar/Postgres_apagándose
=== RUN   TestClasificar/conexión_rota
=== RUN   TestClasificar/red
--- PASS: TestClasificar (0.00s)
    --- PASS: TestClasificar/sin_filas (0.00s)
    --- PASS: TestClasificar/tiempo_agotado (0.00s)
    --- PASS: TestClasificar/check_de_la_tabla (0.00s)
    --- PASS: TestClasificar/Postgres_apagándose (0.00s)
    --- PASS: TestClasificar/conexión_rota (0.00s)
    --- PASS: TestClasificar/red (0.00s)
PASS
ok  	tareas	0.003s
$ go vet ./... && go test -race -count=1 ./...
ok  	tareas	1.020s
?   	tareas/herramientas/bloquear	[no test files]

[demo.sh terminó con código 0]
```

## levantar

```
$ docker compose build -q
 Image gl-curso-tareas:p5 Building 
 Image gl-curso-tareas:p5 Built

$ docker compose up -d --wait
 Network gl-curso-p5_default Creating 
 Network gl-curso-p5_default Created 
 Volume gl-curso-p5_datos Creating 
 Volume gl-curso-p5_datos Created 
 Container gl-curso-p5-db-1 Creating 
 Container gl-curso-p5-db-1 Created 
 Container gl-curso-p5-api-1 Creating 
 Container gl-curso-p5-api-1 Created 
 Container gl-curso-p5-db-1 Starting 
 Container gl-curso-p5-db-1 Started 
 Container gl-curso-p5-db-1 Waiting 
 Container gl-curso-p5-db-1 Healthy 
 Container gl-curso-p5-api-1 Starting 
 Container gl-curso-p5-api-1 Started 
 Container gl-curso-p5-db-1 Waiting 
 Container gl-curso-p5-api-1 Waiting 
 Container gl-curso-p5-db-1 Healthy 
 Container gl-curso-p5-api-1 Healthy

$ docker compose ps -a --format "table {{.Service}}\t{{.Status}}"
SERVICE   STATUS
api       Up 5 seconds (healthy)
db        Up 8 seconds (healthy)

$ curl -s -X POST api:8080/tareas -d '{"titulo":"comprar pan"}'
{"id":1,"titulo":"comprar pan","hecha":false}

$ curl -s -X POST api:8080/tareas -d '{"titulo":"llamar al dentista"}'
{"id":2,"titulo":"llamar al dentista","hecha":false}

$ curl -s -w 'HTTP %{http_code}\n' api:8080/salud
{"bd":"ok","conexiones_abiertas":1,"en_uso":0,"max":4,"esperas":1}
HTTP 200

$ docker inspect gl-curso-p5-api-1 --format "{{.State.Health.Status}}"
healthy

$ docker exec gl-curso-p5-api-1 /tareas salud
salud: 200 {"bd":"ok","conexiones_abiertas":1,"en_uso":0,"max":4,"esperas":1}
```

## bd-caida

```
$ docker compose stop db
 Container gl-curso-p5-db-1 Stopping 
 Container gl-curso-p5-db-1 Stopped

$ curl -s -w 'HTTP %{http_code} en %{time_total} s\n' api:8080/tareas
{"error":"base de datos no disponible"}
HTTP 503 en 0.001273 s

$ curl -s -w 'HTTP %{http_code}\n' -X POST api:8080/tareas -d '{"titulo":"pagar la luz"}'
{"error":"la base de datos no respondió a tiempo"}
HTTP 503

$ curl -s -w 'HTTP %{http_code}\n' api:8080/tareas/abc
{"error":"id inválido: \"abc\""}
HTTP 400

$ curl -s -w 'HTTP %{http_code}\n' api:8080/salud
{"bd":"caida","conexiones_abiertas":2,"en_uso":0,"max":4,"esperas":1}
HTTP 503

[esperé a que Docker marcara la API: unhealthy]

$ docker compose ps -a --format "table {{.Service}}\t{{.Status}}"
SERVICE   STATUS
api       Up 13 seconds (unhealthy)
db        Exited (0) 6 seconds ago

$ docker exec gl-curso-p5-api-1 /tareas salud
salud: 503 {"bd":"caida","conexiones_abiertas":2,"en_uso":0,"max":4,"esperas":1}
[código de salida: 1]

$ docker compose logs api --no-log-prefix --tail 4
{"time":"2026-11-23T12:16:24.952229074Z","level":"INFO","msg":"petición","metodo":"GET","ruta":"/tareas/abc","patron":"GET /tareas/{id}","estado":400,"ms":0}
{"time":"2026-11-23T12:16:26.2909385Z","level":"WARN","msg":"petición","metodo":"GET","ruta":"/salud","patron":"GET /salud","estado":503,"ms":1000}
{"time":"2026-11-23T12:16:28.124183253Z","level":"WARN","msg":"petición","metodo":"GET","ruta":"/salud","patron":"GET /salud","estado":503,"ms":1000}
{"time":"2026-11-23T12:16:29.889869671Z","level":"WARN","msg":"petición","metodo":"GET","ruta":"/salud","patron":"GET /salud","estado":503,"ms":1000}

$ docker compose logs api --no-log-prefix | grep almacén
{"time":"2026-11-23T12:16:22.275258737Z","level":"ERROR","msg":"almacén","err":"base de datos no disponible: FATAL: terminating connection due to administrator command (SQLSTATE 57P01)"}
{"time":"2026-11-23T12:16:24.597452978Z","level":"ERROR","msg":"almacén","err":"la base de datos no respondió a tiempo: context deadline exceeded"}

$ docker inspect gl-curso-p5-api-1 --format "{{.RestartCount}}"
0
```

## bd-vuelve

```
$ docker compose start db
 Container gl-curso-p5-db-1 Starting 
 Container gl-curso-p5-db-1 Started

[esperé a que Docker marcara la API: healthy]

$ docker compose ps -a --format "table {{.Service}}\t{{.Status}}"
SERVICE   STATUS
api       Up 16 seconds (healthy)
db        Up 1 second (health: starting)

$ curl -s -w 'HTTP %{http_code}\n' api:8080/tareas
[{"id":1,"titulo":"comprar pan","hecha":false},{"id":2,"titulo":"llamar al dentista","hecha":false}]
HTTP 200

$ docker inspect gl-curso-p5-api-1 --format "{{.RestartCount}}"
0
```

## api-muere

```
$ docker inspect gl-curso-p5-api-1 --format "{{.State.Pid}}"
27163

$ docker run --rm --pid host alpine:3.24.3 kill -KILL 27163


$ docker compose ps -a --format "table {{.Service}}\t{{.Status}}"
SERVICE   STATUS
api       Up Less than a second (health: starting)
db        Up 3 seconds (healthy)

$ docker inspect gl-curso-p5-api-1 --format "{{.State.ExitCode}} {{.RestartCount}}"
0 1

$ docker compose ps -a --format "table {{.Service}}\t{{.Status}}"
SERVICE   STATUS
api       Up 5 seconds (healthy)
db        Up 7 seconds (healthy)

$ docker inspect gl-curso-p5-api-1 --format "{{.RestartCount}}"
1

$ curl -s -w 'HTTP %{http_code}\n' api:8080/tareas
[{"id":1,"titulo":"comprar pan","hecha":false},{"id":2,"titulo":"llamar al dentista","hecha":false}]
HTTP 200
```

## apagar

```
$ docker compose stop api
 Container gl-curso-p5-api-1 Stopping 
 Container gl-curso-p5-api-1 Stopped

$ docker compose logs api --no-log-prefix --tail 3
{"time":"2026-11-23T12:16:38.393060972Z","level":"INFO","msg":"petición","metodo":"GET","ruta":"/tareas","patron":"GET /tareas","estado":200,"ms":2}
{"time":"2026-11-23T12:16:38.686736046Z","level":"INFO","msg":"señal recibida: apagando"}
{"time":"2026-11-23T12:16:38.686814Z","level":"INFO","msg":"apagado limpio"}

$ docker inspect gl-curso-p5-api-1 --format "{{.State.ExitCode}}"
0

$ docker compose down -v
 Container gl-curso-p5-api-1 Stopping 
 Container gl-curso-p5-api-1 Stopped 
 Container gl-curso-p5-api-1 Removing 
 Container gl-curso-p5-api-1 Removed 
 Container gl-curso-p5-db-1 Stopping 
 Container gl-curso-p5-db-1 Stopped 
 Container gl-curso-p5-db-1 Removing 
 Container gl-curso-p5-db-1 Removed 
 Volume gl-curso-p5_datos Removing 
 Network gl-curso-p5_default Removing 
 Volume gl-curso-p5_datos Removed 
 Network gl-curso-p5_default Removed

$ docker image rm gl-curso-tareas:p5
Untagged: gl-curso-tareas:p5
Deleted: sha256:4b6ab21142866642d03188c6147c05117da409d6fe05a16fab4ddd2a804a881c
```
