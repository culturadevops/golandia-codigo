# SALIDA — Curso de Rust para DevOps 2027

> Generado por `ejecutar_todo.py`. **No se escribe a mano.** Los fallos que se ven son a propósito:
> cada lección enseña la trampa. Los tiempos cambian un poco de una corrida a otra.

- fecha: `2026-11-13 14:10:41Z`
- `rustc 1.99.1 (58cf8f9c3 2026-10-13)`
- `cargo 1.99.1 (5f94df478 2026-08-27)`
- `axum	0.8.9`
- `reqwest	0.13.5`
- `serde_yaml	0.9.34+deprecated`
- `tokio	1.54.1`
- `toml	1.2.1+spec-1.1.0`
- `tracing-subscriber	0.3.23`
- `docker 29.6.2`

## parte1/01_async

código de salida `0`

```
$ tareas hilos 10000
10000 hilos:  1.30 s, memoria pico 94 MB
[codigo de salida: 0]
$ tareas tareas 10000
10000 tareas: 1.01 s, memoria pico 6 MB
[codigo de salida: 0]
$ tareas bloqueo
10 tareas de 100 ms con std::thread::sleep: 1.00 s
[codigo de salida: 0]
$ tareas bien
10 tareas de 100 ms con tokio::time::sleep: 0.10 s
[codigo de salida: 0]
```

## parte1/02_concurrente

código de salida `0`

```
$ salud secuencial
20 endpoints de 200 ms, uno detrás de otro: 4.03 s
[codigo de salida: 0]
$ salud concurrente
20 endpoints a la vez con join_all: 0.20 s (20 respuestas)
[codigo de salida: 0]
$ salud limite
20 endpoints, de 5 en 5 con buffer_unordered: 0.81 s (20 respuestas)
[codigo de salida: 0]
```

## parte1/03_timeouts

código de salida `0`

```
$ salud sin-timeout
sin timeout: Ok(200) tras 5.00 s
[codigo de salida: 0]
$ salud con-timeout
con timeout: error tras 1.00 s, is_timeout=true: error sending request for url (http://127.0.0.1:8080/lento?ms=5000)
[codigo de salida: 0]
$ salud ingenuo
ARRIBA /ok
ARRIBA /falla
[codigo de salida: 0]
$ salud estricto
ARRIBA /ok (200 OK)
CAIDO  /falla: HTTP status server error (500 Internal Server Error) for url (http://127.0.0.1:8080/falla)
[codigo de salida: 0]
```

## parte1/04_reintentos

código de salida `0`

```
$ salud sin-reintentos
un intento: 503 Service Unavailable
[codigo de salida: 0]
$ salud reintentos
intento 1: 503 Service Unavailable · espero 100 ms
intento 2: 503 Service Unavailable · espero 200 ms
intento 3: 200 OK · total 0.30 s
[codigo de salida: 0]
```

## parte2/01_yaml

código de salida `0`

```
--- buena.yaml
checks:
  - nombre: api
    url: https://api.ejemplo.com/health
    intervalo_s: 30
    timeout_ms: 1500
  - nombre: web
    url: https://www.ejemplo.com
    intervalo_s: 60
---
$ config leer buena.yaml
api -> https://api.ejemplo.com/health cada 30 s, timeout 1500 ms
web -> https://www.ejemplo.com cada 60 s, timeout 2000 ms
[codigo de salida: 0]
--- mal_indentado.yaml
checks:
  - nombre: api
    url: https://api.ejemplo.com/health
   intervalo_s: 30
---
$ config leer mal_indentado.yaml
error: mal_indentado.yaml: checks[0]: missing field `intervalo_s` at line 2 column 5
[codigo de salida: 1]
--- comillas.yaml
checks:
  - nombre: api
    url: https://api.ejemplo.com/health
    intervalo_s: "30"
---
$ config leer comillas.yaml
error: comillas.yaml: checks[0].intervalo_s: invalid type: string "30", expected u64 at line 4 column 18
[codigo de salida: 1]
--- typo.yaml
checks:
  - nombre: api
    url: https://api.ejemplo.com/health
    intervalo_s: 30
    timout_ms: 500
---
$ config leer typo.yaml
api -> https://api.ejemplo.com/health cada 30 s, timeout 2000 ms
[codigo de salida: 0]
$ config estricto typo.yaml
error: typo.yaml: checks[0]: unknown field `timout_ms`, expected one of `nombre`, `url`, `timeout_ms`, `intervalo_s` at line 5 column 5
[codigo de salida: 1]
```

## parte2/02_toml

código de salida `0`

```
--- buena.toml
[[checks]]
nombre = "api"
url = "https://api.ejemplo.com/health"
intervalo_s = 30
timeout_ms = 1500

[[checks]]
nombre = "web"
url = "https://www.ejemplo.com"
intervalo_s = 60
---
$ config leer buena.toml
api -> https://api.ejemplo.com/health cada 30 s, timeout 1500 ms
web -> https://www.ejemplo.com cada 60 s, timeout 2000 ms
[codigo de salida: 0]
--- mal.toml
[[checks]]
nombre = "api"
url = https://api.ejemplo.com/health
intervalo_s = 30
---
$ config leer mal.toml
error: mal.toml: TOML parse error at line 3, column 7
  |
3 | url = https://api.ejemplo.com/health
  |       ^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^
string values must be quoted, expected literal string

[codigo de salida: 1]
```

## parte2/03_validar

código de salida `0`

```
--- invalida.yaml
checks:
  - nombre: api
    url: api.ejemplo.com/health
    intervalo_s: 30
  - nombre: web
    url: https://www.ejemplo.com
    intervalo_s: 0
  - nombre: lenta
    url: https://lenta.ejemplo.com
    intervalo_s: 2
    timeout_ms: 5000
---
$ config leer invalida.yaml
api -> api.ejemplo.com/health cada 30 s, timeout 2000 ms
web -> https://www.ejemplo.com cada 0 s, timeout 2000 ms
lenta -> https://lenta.ejemplo.com cada 2 s, timeout 5000 ms
[codigo de salida: 0]
$ config validar invalida.yaml
error: api: la url debe empezar por http:// o https:// (es "api.ejemplo.com/health")
error: web: intervalo_s no puede ser 0
error: web: timeout_ms (2000) debe ser menor que el intervalo (0 s)
error: lenta: timeout_ms (5000) debe ser menor que el intervalo (2 s)
4 errores
[codigo de salida: 1]
$ config validar ../01_yaml/buena.yaml
ok: 2 checks válidos
[codigo de salida: 0]
```

## parte2/04_comandos

código de salida `0`

```
$ comandos ingenuo
ls terminó bien (¿seguro?)
[codigo de salida: 0]
$ comandos estricto
éxito: false
código: Some(2)
stderr: ls: cannot access '/no-existe': No such file or directory
[codigo de salida: 0]
$ comandos no-instalado
no se pudo lanzar kubectl: No such file or directory (os error 2) (kind: NotFound)
[codigo de salida: 0]
$ comandos sin-shell
sin shell:  hola; echo HACKEADO
con sh -c:  hola | HACKEADO | 
[codigo de salida: 0]
```

## parte2/05_timeout_comando

código de salida `0`

```
$ comandos timeout
timeout tras 1.00 s (timeout)
[codigo de salida: 0]
$ pgrep -a sleep
12767 sleep 30
$ comandos timeout-kill
timeout tras 1.00 s (timeout-kill)
[codigo de salida: 0]
$ pgrep -a sleep
(ninguno)
```

## parte3/01_metricas

código de salida `0`

```
$ curl -s localhost:8090/health
ok
$ curl -s localhost:8090/metrics
# HELP rdevops_peticiones_total Peticiones atendidas por ruta.
# TYPE rdevops_peticiones_total counter
rdevops_peticiones_total{ruta="/health"} 4
rdevops_peticiones_total{ruta="/lento"} 1
rdevops_peticiones_total{ruta="/metrics"} 1
```

## parte3/02_logs

código de salida `0`

```
$ cat servicio.log (los timestamps, recortados)
{"level":"INFO","fields":{"message":"escuchando","puerto":8090,"sin_apagado":false}}
{"level":"INFO","fields":{"message":"petición lenta empieza","ms":50}}
{"level":"INFO","fields":{"message":"petición lenta termina","ms":50}}
{"level":"INFO","fields":{"message":"SIGTERM recibido: termino lo que está en vuelo"}}
{"level":"INFO","fields":{"message":"apagado limpio"}}
$ grep -c lenta servicio.log
2
```

## parte3/03_apagado

código de salida `0`

```
$ servicio --sin-apagado   # y a 1 s de una petición de 3 s, kill -TERM
  respuesta del cliente:
    curl: (52) Empty reply from server
    [curl: código 52]
  el servicio salió con código 143 a los 1.0 s
  últimas líneas del log:
    {"level":"INFO","fields":{"message":"escuchando","puerto":8090,"sin_apagado":true}}
    {"level":"INFO","fields":{"message":"petición lenta empieza","ms":3000}}
$ servicio    # y a 1 s de una petición de 3 s, kill -TERM
  respuesta del cliente:
    listo tras 3000 ms
    [curl: código 0]
  el servicio salió con código 0 a los 3.0 s
  últimas líneas del log:
    {"level":"INFO","fields":{"message":"petición lenta termina","ms":3000}}
    {"level":"INFO","fields":{"message":"apagado limpio"}}
```

## parte3/04_imagen

código de salida `0`

```
$ docker build -f parte3/04_imagen/Dockerfile.glibc -t rdevops-servicio:glibc .
listo
$ docker build -f parte3/04_imagen/Dockerfile.musl -t rdevops-servicio:musl .
listo
$ docker image inspect -f '{{.Size}}' rdevops-servicio:glibc rdevops-servicio:musl
glibc: 30740200 bytes (29.3 MiB)
musl: 959941 bytes (0.9 MiB)
$ docker run -d -p 18090:8090 rdevops-servicio:musl && curl localhost:18090/health
ok
```

## parte3/05_pid1

código de salida `0`

```
# sin manejar SIGTERM
$ docker run -d --name rdevops-pid1 rdevops-servicio:musl --sin-apagado
$ time docker stop rdevops-pid1
[tardó 10.5 s]
código de salida del contenedor: 137
última línea del log: {"level":"INFO","fields":{"message":"escuchando","puerto":8090,"sin_apagado":true}}
# con apagado limpio
$ docker run -d --name rdevops-pid1 rdevops-servicio:musl
$ time docker stop rdevops-pid1
[tardó 0.2 s]
código de salida del contenedor: 0
última línea del log: {"level":"INFO","fields":{"message":"apagado limpio"}}
```
