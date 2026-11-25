# SALIDA — Go práctico 2027 · parte 5

> Generado por `ejecutar_todo.py`. **No se escribe a mano.** Ids, tiempos y tamaños son los de esta corrida.

- fecha: `2026-11-23 12:13:12Z`
- imagen Go: `golang:1.27.2` · Postgres: `postgres:18.6-alpine` · `Docker 29.6.2`

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
db        Up 10 seconds (healthy)

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
{"error":"error interno"}
HTTP 500 en 0.001327 s

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
{"time":"2026-11-23T12:12:57.933021304Z","level":"INFO","msg":"petición","metodo":"GET","ruta":"/tareas/abc","patron":"GET /tareas/{id}","estado":400,"ms":0}
{"time":"2026-11-23T12:12:59.268111913Z","level":"WARN","msg":"petición","metodo":"GET","ruta":"/salud","patron":"GET /salud","estado":503,"ms":1001}
{"time":"2026-11-23T12:13:00.955490827Z","level":"WARN","msg":"petición","metodo":"GET","ruta":"/salud","patron":"GET /salud","estado":503,"ms":1000}
{"time":"2026-11-23T12:13:02.338124038Z","level":"WARN","msg":"petición","metodo":"GET","ruta":"/salud","patron":"GET /salud","estado":503,"ms":1001}

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
16226

$ docker run --rm --pid host alpine:3.24.3 kill -KILL 16226


$ docker compose ps -a --format "table {{.Service}}\t{{.Status}}"
SERVICE   STATUS
api       Up 1 second (health: starting)
db        Up 3 seconds (healthy)

$ docker inspect gl-curso-p5-api-1 --format "{{.State.ExitCode}} {{.RestartCount}}"
0 1

$ docker compose ps -a --format "table {{.Service}}\t{{.Status}}"
SERVICE   STATUS
api       Up 5 seconds (healthy)
db        Up 8 seconds (healthy)

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
{"time":"2026-11-23T12:13:11.378104493Z","level":"INFO","msg":"petición","metodo":"GET","ruta":"/tareas","patron":"GET /tareas","estado":200,"ms":1}
{"time":"2026-11-23T12:13:11.631807057Z","level":"INFO","msg":"señal recibida: apagando"}
{"time":"2026-11-23T12:13:11.631888076Z","level":"INFO","msg":"apagado limpio"}

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
 Network gl-curso-p5_default Removing 
 Volume gl-curso-p5_datos Removing 
 Volume gl-curso-p5_datos Removed 
 Network gl-curso-p5_default Removed

$ docker image rm gl-curso-tareas:p5
Untagged: gl-curso-tareas:p5
Deleted: sha256:2473d99ffc957293f09c66ee1c78f0327203571923594951390d290a0d975510
```
