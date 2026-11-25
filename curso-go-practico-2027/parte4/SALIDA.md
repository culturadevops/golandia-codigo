# SALIDA — Go práctico 2027 · parte 4

> Generado por `ejecutar_todo.py`. **No se escribe a mano.** Ids, tiempos y tamaños son los de esta corrida.

- fecha: `2026-11-23 12:03:46Z`
- imagen Go: `golang:1.27.2` · Postgres: `postgres:18.6-alpine` · `Docker 29.6.2`

## config

```
$ env -u DATABASE_URL ./tareas
{"time":"2026-11-23T12:03:32.25033864Z","level":"ERROR","msg":"la API no arrancó","err":"configuración: DATABASE_URL es obligatoria"}
[código de salida: 1]
$ env -u DATABASE_URL TAREAS_PUERTO=ochenta TAREAS_TIMEOUT_BD=2 ./tareas
{"time":"2026-11-23T12:03:32.254329689Z","level":"ERROR","msg":"la API no arrancó","err":"configuración: TAREAS_PUERTO=\"ochenta\": tiene que ser un número de 1 a 65535\nDATABASE_URL es obligatoria\nTAREAS_TIMEOUT_BD=\"2\": usa una duración como 2s o 500ms"}
[código de salida: 1]
$ go test -run LeerConfig -v .
=== RUN   TestLeerConfigValoresPorDefecto
--- PASS: TestLeerConfigValoresPorDefecto (0.00s)
=== RUN   TestLeerConfigJuntaLosErrores
--- PASS: TestLeerConfigJuntaLosErrores (0.00s)
PASS
ok  	tareas	(cached)
```

## logs

```
$ ./tareas &
$ curl -s -X POST localhost:8080/tareas -d '{"titulo":"comprar pan"}'
{"id":1,"titulo":"comprar pan","hecha":false}
$ curl -s -o /dev/null -w 'HTTP %{http_code}\n' localhost:8080/tareas/99
HTTP 404
$ curl -s -o /dev/null -w 'HTTP %{http_code}\n' -X POST localhost:8080/tareas -d '{"titulo":""}'
HTTP 422
$ cat api.log
{"time":"2026-11-23T12:03:32.407076455Z","level":"INFO","msg":"migraciones","aplicadas":["001_crear_tareas.sql","002_fecha_de_creacion.sql"]}
{"time":"2026-11-23T12:03:32.407197018Z","level":"INFO","msg":"escuchando","puerto":8080,"timeout_bd":"5s"}
{"time":"2026-11-23T12:03:32.497412687Z","level":"INFO","msg":"petición","metodo":"POST","ruta":"/tareas","patron":"POST /tareas","estado":201,"ms":2}
{"time":"2026-11-23T12:03:32.50357769Z","level":"INFO","msg":"petición","metodo":"GET","ruta":"/tareas/99","patron":"GET /tareas/{id}","estado":404,"ms":0}
{"time":"2026-11-23T12:03:32.508781775Z","level":"INFO","msg":"petición","metodo":"POST","ruta":"/tareas","patron":"POST /tareas","estado":422,"ms":0}
$ kill -TERM $PID
[la API terminó con código 0]
```

## sin-apagado

```
$ ./tareas-parte2 &
$ ./bloquear 3 &
$ curl -s -w 'HTTP %{http_code}\n' localhost:8080/tareas &
$ kill -TERM $PID
[curl terminó con código 52]
HTTP 000
[la API terminó con código 143]
bloquear: tabla tareas bloqueada 3 s
bloquear: tabla tareas liberada
```

## apagado

```
$ ./tareas &
$ ./bloquear 3 &
$ curl -s -w 'HTTP %{http_code}\n' localhost:8080/tareas &
$ kill -TERM $PID
$ curl -s -w 'HTTP %{http_code}\n' localhost:8080/tareas
HTTP 000
[código de salida: 7]
[{"id":1,"titulo":"comprar pan","hecha":false}]
HTTP 200
[la API terminó con código 0]
bloquear: tabla tareas bloqueada 3 s
bloquear: tabla tareas liberada
$ cat api.log
{"time":"2026-11-23T12:03:37.405935678Z","level":"INFO","msg":"señal recibida: apagando"}
{"time":"2026-11-23T12:03:39.408000497Z","level":"INFO","msg":"petición","metodo":"GET","ruta":"/tareas","patron":"GET /tareas","estado":200,"ms":2492}
{"time":"2026-11-23T12:03:39.557605245Z","level":"INFO","msg":"apagado limpio"}

[demo.sh terminó con código 0]
```

## imagenes

```
$ docker build -q -f Dockerfile.una-etapa -t gl-curso-tareas:una-etapa .
sha256:443751d084227283e6315342b65e9d0eaeedd78a1e5a4e099537e6e48fc056f1

$ docker build -q -t gl-curso-tareas:p4 .
sha256:a1f5aa55b23daf9b39e0490c3305d8a8bf24f76da455ad27cd3518c914a79dc7

$ docker image inspect gl-curso-tareas:una-etapa --format "{{.Size}} bytes · usuario: '{{.Config.User}}'"
372356781 bytes · usuario: ''

$ docker image inspect gl-curso-tareas:p4 --format "{{.Size}} bytes · usuario: '{{.Config.User}}'"
4157886 bytes · usuario: '65532:65532'

$ docker image ls gl-curso-tareas --format "{{.Tag}}  {{.Size}}"
p4  15.4MB
una-etapa  1.56GB

$ docker history gl-curso-tareas:p4 --format "{{.Size}}  {{.CreatedBy}}"
0B  ENTRYPOINT ["/tareas"]
0B  EXPOSE [8080/tcp]
0B  USER 65532:65532
11.2MB  COPY /tareas /tareas # buildkit

$ docker run --rm gl-curso-tareas:p4
{"time":"2026-11-23T12:03:41.929305035Z","level":"ERROR","msg":"la API no arrancó","err":"configuración: DATABASE_URL es obligatoria"}
[código de salida: 1]

$ docker run --rm --entrypoint sh gl-curso-tareas:p4
docker: Error response from daemon: failed to create task for container: failed to create shim task: OCI runtime create failed: runc create failed: unable to start container process: error during container init: exec: "sh": executable file not found in $PATH

Run 'docker run --help' for more information
[código de salida: 127]
```

## contenedor

```
$ docker run -d --rm --name gl-curso-api-p4 --network gl-curso-red -e DATABASE_URL=... gl-curso-tareas:p4

$ curl -s -X POST gl-curso-api-p4:8080/tareas -d '{"titulo":"desde Docker"}'
{"id":1,"titulo":"desde Docker","hecha":false}

$ docker stop gl-curso-api-p4
gl-curso-api-p4

$ docker logs -f gl-curso-api-p4
{"time":"2026-11-23T12:03:44.058090173Z","level":"INFO","msg":"migraciones","aplicadas":["001_crear_tareas.sql","002_fecha_de_creacion.sql"]}
{"time":"2026-11-23T12:03:44.058367416Z","level":"INFO","msg":"escuchando","puerto":8080,"timeout_bd":"2s"}
{"time":"2026-11-23T12:03:44.276911944Z","level":"INFO","msg":"petición","metodo":"GET","ruta":"/salud","patron":"GET /salud","estado":200,"ms":0}
{"time":"2026-11-23T12:03:44.653582541Z","level":"INFO","msg":"petición","metodo":"POST","ruta":"/tareas","patron":"POST /tareas","estado":201,"ms":2}
{"time":"2026-11-23T12:03:45.858768097Z","level":"INFO","msg":"señal recibida: apagando"}
{"time":"2026-11-23T12:03:45.85890778Z","level":"INFO","msg":"apagado limpio"}

[docker wait: código de salida 0]
```
