# SALIDA — curso «Una API en Go de cero a producción» (golandia 685)

> Generado por `python ejecutar_todo.py`. **No se escribe a mano.**

- fecha: `2026-10-28 18:21 UTC` · sistema anfitrión: `Windows 10`
- imágenes: `golang:1.27.2`, `postgres:18.6-alpine`, `alpine:3.24.3`, act `v0.2.90` con `catthehacker/ubuntu:act-latest`
- los tiempos dependen de la máquina: Docker Desktop sobre Windows

## Versiones

```
$ docker version --format '{{.Server.Version}}'
29.6.2
[código de salida 0]
```

```
$ go version
go version go1.27.2 linux/amd64
[código de salida 0]
```

```
$ postgres --version
postgres (PostgreSQL) 18.6
[código de salida 0]
```

```
$ go list -m github.com/jackc/pgx/v5 github.com/pressly/goose/v3 golang.org/x/crypto
github.com/jackc/pgx/v5 v5.11.0
github.com/pressly/goose/v3 v3.29.0
golang.org/x/crypto v0.58.0
[código de salida 0]
```

## Lección 1 — La estructura: main solo llama a run()

La plantilla (sin `lecciones/`), con las líneas de cada archivo:

```
.dockerignore  (8)
.gitignore  (14)
Dockerfile  (24)
README.md  (142)
docker-compose.yml  (46)
go.mod  (20)
go.sum  (53)
.github/
  workflows/
    ci.yml  (46)
cmd/
  api/
    main.go  (42)
    ordenes.go  (75)
internal/
  bd/
    bd.go  (77)
  config/
    config.go  (111)
    config_test.go  (47)
  cuentas/
    cuentas.go  (126)
    cuentas_test.go  (58)
    http.go  (93)
  pruebabd/
    pruebabd.go  (71)
  servidor/
    salud.go  (69)
    servidor.go  (86)
    servidor_test.go  (72)
  tareas/
    http.go  (135)
    http_test.go  (90)
    postgres.go  (59)
    postgres_test.go  (55)
    tareas.go  (95)
  web/
    errores.go  (82)
    leer.go  (53)
    middleware.go  (153)
    mux.go  (42)
    web_test.go  (88)
migraciones/
  00001_usuarios.sql  (22)
  00002_tareas.sql  (14)
  migraciones.go  (26)
sql/
  roles.sql  (21)
```

```
$ go run ./lecciones/01_estructura -modo fatal
  abro el pool de la base de datos
  no puedo escuchar en :8080
exit status 1
[código de salida 1]
```

```
$ go run ./lecciones/01_estructura -modo run
  abro el pool de la base de datos
  defer: cierro el pool y vacío los logs
main: error: no puedo escuchar en :8080
exit status 1
[código de salida 1]
```

## Lección 2 — El router de net/http (Go 1.22+)

```
$ go run ./lecciones/02_router
antes de 1.22 ("/" y "/tareas/"):
  GET /tareas/7        -> 200  tarea 7
  DELETE /tareas/7     -> 200  tarea 7
  GET /tarea           -> 200  inicio
  GET /no-existe       -> 200  inicio

Go 1.22+ ("GET /{$}", "GET /tareas/{id}"):
  GET /tareas/7        -> 200  tarea 7
  POST /tareas/7       -> 405  Method Not Allowed  Allow: DELETE, GET, HEAD
  GET /tarea           -> 404  404 page not found
  GET /tareas/7/x      -> 404  404 page not found

lo mismo + un "/" para el 404 en JSON:
  POST /tareas/7       -> 404  {"error":"no existe"}

el de la plantilla (ErroresDelMux):
  DELETE /salud/vivo   -> 405  {"error":{"codigo":"metodo_no_permitido","mensaje":"permitidos: GET, HEAD"},"request_id":"4892910986926e4d"}  Allow: GET, HEAD
  GET /no-existe       -> 404  {"error":{"codigo":"no_encontrado","mensaje":"no existe esa ruta"},"request_id":"943cac85cbe41050"}
[código de salida 0]
```

```
$ GODEBUG=httpmuxgo121=1 go run ./lecciones/02_router
GODEBUG=httpmuxgo121=1
patrones de 1.22 con el mux viejo:
  GET /tareas/7        -> 404  404 page not found
  GET /                -> 404  404 page not found
[código de salida 0]
```

## Lección 3 — Middlewares: request id, registro con slog, recover y plazo

```
$ go run ./lecciones/03_middlewares
1) un handler que solo hace Write (Go manda 200 solo):
  ingenuo:   level=INFO msg=peticion ruta=/tareas estado=0
  plantilla: level=INFO msg=peticion request_id=… metodo=GET ruta=/tareas estado=200 bytes=2 ms=0

2) un handler que entra en pánico:
  sin Recuperar, el cliente recibe: Get "http://127.0.0.1:37143/tareas/1": EOF
  y el servidor escribe: http: panic serving 127.0.0.1:55878: índice fuera de rango
  con Recuperar: 500 {"error":{"codigo":"interno","mensaje":"error interno"},"request_id":"dd3d4e60e3d1b7ac"}
    log: level=ERROR msg=panico request_id=dd3d4e60e3d1b7ac valor="índice fuera de rango" pila="(…29 líneas)"
    log: level=INFO msg=peticion request_id=dd3d4e60e3d1b7ac metodo=GET ruta=/tareas/1 estado=500 bytes=89

3) plazo de 250 ms; el handler hace 10 pasos de 100 ms:
  ignora r.Context() -> 200 en 1000 ms, 10 de 10 pasos hechos  {"pasos":10}
  mira r.Context()   -> 503 en  250 ms,  2 de 10 pasos hechos  {"error":{"codigo":"tiempo_agotado","mensaje":"la petición tardó demasiado"}}
[código de salida 0]
```

## Lección 4 — Configuración por entorno, validada al arrancar

```
$ go run ./lecciones/04_config
1) os.Getenv sin validar, con PUERTO sin definir:
  net.Listen(":"+"") -> escuchando en [::]:39557, err = <nil>

2) TIEMPO_PETICION=30 (sin unidad), con el error ignorado:
  plazo = 0s  -> sin plazo: una petición colgada espera para siempre

3) registrar la configuración al arrancar:
  string:      level=INFO msg=arrancando bd=postgres://app:S3cr3ta-de-verdad@bd:5432/tareas
  config.DSN:  level=INFO msg=arrancando bd=postgres://app:xxxxx@bd:5432/tareas

4) config.Cargar con el entorno mal en cuatro sitios:
PUERTO: "80a" no es un puerto (1-65535)
BD_URL: obligatorio
TIEMPO_PETICION: "30" no es una duración (ejemplo: 30s, 500ms)
NIVEL_LOG: "todo" no es debug, info, warn ni error
[código de salida 0]
```

El binario de verdad, con el entorno mal:

```
$ PUERTO=80a TIEMPO_PETICION=30 go run ./cmd/api
error: configuración:
PUERTO: "80a" no es un puerto (1-65535)
BD_URL: obligatorio
TIEMPO_PETICION: "30" no es una duración (ejemplo: 30s, 500ms)
[código de salida 1]
```

## Lección 5 — Postgres: migraciones embebidas y un usuario sin DDL

```
$ psql -U postgres -f sql/roles.sql   (lo hizo el contenedor al crearse; así quedó)
 rolname  | rolsuper | rolcreatedb 
----------+----------+-------------
 app      | f        | f
 migrador | f        | f
(2 rows)
[código de salida 0]
```

```
$ BD_URL_MIGRAR=…migrador… api migrar
aplicada 00001_usuarios.sql (6 ms)
aplicada 00002_tareas.sql (3 ms)
esquema en la versión 2 (2 migraciones nuevas)
[código de salida 0]
```

```
$ BD_URL_MIGRAR=…migrador… api migrar   (otra vez)
esquema en la versión 2 (0 migraciones nuevas)
[código de salida 0]
```

Con el usuario de la app, que es con el que corre el servidor. Sobre `tareas` no hay nada que aplicar y sale bien; sobre una base vacía (`vacia`, de migrador) tiene que crear tablas:

```
$ BD_URL_MIGRAR=…app…/tareas api migrar
esquema en la versión 2 (0 migraciones nuevas)
[código de salida 0]
```

```
$ BD_URL_MIGRAR=…app…/vacia api migrar
error: failed to initialize: create version table: failed to create version table "goose_db_version": ERROR: permission denied for schema public (SQLSTATE 42501)
exit status 1
[código de salida 1]
```

```
$ psql -U app -c 'DROP TABLE tareas'
ERROR:  must be owner of table tareas
[código de salida 1]
```

```
$ psql -U app -c 'CREATE TABLE copia (x int)'
ERROR:  permission denied for schema public
LINE 1: CREATE TABLE copia (x int)
                     ^
[código de salida 1]
```

```
$ psql -U app -c 'TRUNCATE tareas'
ERROR:  permission denied for table tareas
[código de salida 1]
```

Editar una migración ya aplicada (base aparte, `demo05`):

```
$ go run ./lecciones/05_migraciones
1) se aplica 00001_notas.sql:
  Up -> 1 migraciones aplicadas, versión 1, err = <nil>

2) alguien EDITA 00001_notas.sql para añadir la columna texto, y vuelve a migrar:
  Up -> 0 migraciones aplicadas, versión 1, err = <nil>
  INSERT con texto -> ERROR: column "texto" of relation "notas" does not exist (SQLSTATE 42703)

3) lo correcto: 00001 se queda como estaba y el cambio va en 00002:
  Up -> 1 migraciones aplicadas, versión 2, err = <nil>
  INSERT con texto -> ok
[código de salida 0]
```

## Lección 6 — Autenticación: sesiones en Postgres y bcrypt

```
$ go run ./lecciones/06_auth
1) lo que tarda bcrypt en cada costo (mediana de 3):
  costo 10 -> 35.6 ms
  costo 11 -> 70.1 ms
  costo 12 -> 139.9 ms
  costo 13 -> 291.8 ms
  costo 14 -> 559.4 ms

2) iniciar sesión mal, 15 veces cada caso (mediana):
  ingenuo   correo que NO existe: 0.1 ms
  ingenuo   correo que SÍ existe: 142.2 ms
  plantilla correo que NO existe: 143.1 ms
  plantilla correo que SÍ existe: 140.6 ms

3) Beto pide la tarea de Ana por su id:
  ingenuo   WHERE id = $1                  -> "pagar la renta"
  plantilla WHERE id = $1 AND usuario_id = $2 -> no_encontrado: esa tarea no existe

4) la sesión:
  token al cliente: dsaN5ncIWrkf… (43 caracteres)
  en la tabla:      e370c1d3f6e2… (SHA-256, 32 bytes)
  buscar la sesión en cada petición: 0.1 ms (mediana de 200)
  tras cerrar sesión, el mismo token -> no_autenticado: sesión inválida o vencida
[código de salida 0]
```

## Lección 7 — Validación y errores JSON coherentes

```
$ go run ./lecciones/07_errores
campo mal escrito (priorida):
  ingenuo   -> 201 {"id":2,"titulo":"llamar al banco","prioridad":2,"hecha":false,"creada":"…"}
  plantilla -> 400 {"error":{"codigo":"peticion_invalida","mensaje":"campo desconocido \"priorida\""},"request_id":"…"}
dos objetos seguidos:
  ingenuo   -> 201 {"id":3,"titulo":"uno","prioridad":2,"hecha":false,"creada":"…"}
  plantilla -> 400 {"error":{"codigo":"peticion_invalida","mensaje":"el cuerpo debe tener un solo objeto JSON"},"request_id":"…"}
cuerpo de 5 MB:
  ingenuo   -> 201 {"id":4,"titulo":"grande","prioridad":2,"hecha":false,"creada":"…"}
  plantilla -> 413 {"error":{"codigo":"demasiado_grande","mensaje":"el cuerpo pasa de 65536 bytes"},"request_id":"…"}
título vacío y prioridad 9:
  ingenuo   -> 500 ERROR: new row for relation "tareas" violates check constraint "tareas_prioridad_check" (SQLSTATE 23514)
  plantilla -> 422 {"error":{"codigo":"validacion","mensaje":"hay campos con problemas","campos":{"prioridad":"1, 2 o 3","titulo":"obligatorio"}},"request_id":"…"}
correo ya registrado:
  ingenuo   -> 500 ERROR: duplicate key value violates unique constraint "usuarios_email" (SQLSTATE 23505)
  plantilla -> 409 {"error":{"codigo":"conflicto","mensaje":"ese correo ya está registrado"},"request_id":"…"}
[código de salida 0]
```

## Lección 8 — Tests unitarios y de integración contra Postgres

Sin las variables de la base:

```
$ go test ./...
?   	tareas/internal/bd	[no test files]
ok  	tareas/internal/config	0.003s
ok  	tareas/internal/cuentas	0.003s
?   	tareas/internal/pruebabd	[no test files]
ok  	tareas/internal/servidor	0.004s
ok  	tareas/internal/tareas	0.004s
ok  	tareas/internal/web	0.003s
?   	tareas/cmd/api	[no test files]
[código de salida 0]
```

```
$ go test -v ./... | grep -E 'SKIP|sin TEST_BD_URL'
    cuentas_test.go:25: sin TEST_BD_URL: test de integración saltado
--- SKIP: TestCuentas (0.00s)
    servidor_test.go:23: sin TEST_BD_URL: test de integración saltado
--- SKIP: TestDePuntaAPunta (0.00s)
    postgres_test.go:14: sin TEST_BD_URL: test de integración saltado
--- SKIP: TestPostgres (0.00s)
[código de salida 0]
```

```
$ CI=true go test ./...
ok  	tareas/internal/config	(cached)
--- FAIL: TestCuentas (0.00s)
    cuentas_test.go:25: CI sin TEST_BD_URL / TEST_BD_URL_MIGRAR: los tests de integración no pueden saltarse
FAIL
FAIL	tareas/internal/cuentas	0.004s
--- FAIL: TestDePuntaAPunta (0.00s)
    servidor_test.go:23: CI sin TEST_BD_URL / TEST_BD_URL_MIGRAR: los tests de integración no pueden saltarse
FAIL
FAIL	tareas/internal/servidor	0.003s
--- FAIL: TestPostgres (0.00s)
    postgres_test.go:14: CI sin TEST_BD_URL / TEST_BD_URL_MIGRAR: los tests de integración no pueden saltarse
FAIL
FAIL	tareas/internal/tareas	0.004s
ok  	tareas/internal/web	(cached)
FAIL
[código de salida 0]
```

Con la base de tests:

```
$ go test ./...
ok  	tareas/internal/config	(cached)
ok  	tareas/internal/cuentas	0.275s
ok  	tareas/internal/servidor	0.390s
ok  	tareas/internal/tareas	0.536s
ok  	tareas/internal/web	(cached)
[código de salida 0]
```

```
$ go test ./...   (otra vez)
ok  	tareas/internal/config	(cached)
ok  	tareas/internal/cuentas	(cached)
ok  	tareas/internal/servidor	(cached)
ok  	tareas/internal/tareas	(cached)
ok  	tareas/internal/web	(cached)
[código de salida 0]
```

Ahora se rompe la base de tests: la columna titulo pasa a llamarse nombre.

```
$ psql -U migrador -d tareas_test -c 'ALTER TABLE tareas RENAME COLUMN titulo TO nombre'
ALTER TABLE
[código de salida 0]
```

```
$ go test ./...
ok  	tareas/internal/config	(cached)
ok  	tareas/internal/cuentas	(cached)
ok  	tareas/internal/servidor	(cached)
ok  	tareas/internal/tareas	(cached)
ok  	tareas/internal/web	(cached)
[código de salida 0]
```

```
$ go test -count=1 ./...
ok  	tareas/internal/config	0.003s
ok  	tareas/internal/cuentas	0.279s
--- FAIL: TestDePuntaAPunta (0.39s)
FAIL
FAIL	tareas/internal/servidor	0.399s
--- FAIL: TestPostgres (0.53s)
    postgres_test.go:27: ERROR: column "titulo" of relation "tareas" does not exist (SQLSTATE 42703)
FAIL
FAIL	tareas/internal/tareas	0.533s
ok  	tareas/internal/web	0.004s
FAIL
[código de salida 0]
```

Base arreglada otra vez, y con el detector de carreras:

```
$ go test -race -count=1 ./...
ok  	tareas/internal/config	1.011s
ok  	tareas/internal/cuentas	3.922s
ok  	tareas/internal/servidor	6.875s
ok  	tareas/internal/tareas	5.404s
ok  	tareas/internal/web	1.016s
[código de salida 0]
```

Funciones de test: 8 unitarias y 3 de integración (las que llaman a `pruebabd.Abrir`).

## Lección 9 — Dockerfile multietapa, medido

```
$ docker build -f lecciones/09_docker/Dockerfile.ingenuo -t g685-tareas:ingenuo .
(construida en 1 s)
[código de salida 0]
```

```
$ docker build -f Dockerfile -t g685-tareas:final .
(construida en 1 s)
[código de salida 0]
```

```
$ docker build -f lecciones/09_docker/Dockerfile.con-cgo -t g685-tareas:con-cgo .
(construida en 1 s)
[código de salida 0]
```

```
$ docker image ls g685-tareas
g685-tareas:con-cgo  17.2MB
g685-tareas:final  17.5MB
g685-tareas:ingenuo  1.6GB
[código de salida 0]
```

```
$ docker inspect -f 'usuario={{.Config.User}}' g685-tareas:ingenuo
usuario=
[código de salida 0]
```

```
$ docker inspect -f 'usuario={{.Config.User}}' g685-tareas:final
usuario=65532:65532
[código de salida 0]
```

```
$ docker run --rm g685-tareas:ingenuo id
uid=0(root) gid=0(root) groups=0(root)
[código de salida 0]
```

```
$ docker run --rm g685-tareas:con-cgo
exec /api: no such file or directory
[código de salida 255]
```

```
$ docker run --rm g685-tareas:final   (sin entorno)
error: configuración:
BD_URL: obligatorio
[código de salida 1]
```

## Lección 10 — Healthcheck: vivo y listo

Contenedor `g685-api` (imagen final). Docker lo marcó healthy a los 5 s.

```
$ curl /salud/vivo ; curl /salud/listo
{"estado":"vivo"}
  [200 en 0.001066 s]
{"esperado":2,"esquema":2,"listo":true,"motivo":""}
  [200 en 0.001239 s]
[código de salida 0]
```

Se pausa Postgres (`docker pause`): no se cae, deja de contestar. Lo peor para una sonda.

```
$ docker pause g685-pg; curl /salud/vivo ; curl /salud/listo
{"estado":"vivo"}
  [200 en 0.001290 s]
{"esperado":2,"esquema":0,"listo":false,"motivo":"base de datos: context deadline exceeded"}
  [503 en 1.002047 s]
[código de salida 0]
```

```
$ go run ./lecciones/10_salud   (con Postgres pausado)
  Ping(context.Background()) -> sin respuesta a los 5 s (la sonda se queda colgada)
  Ping con plazo de 1 s      -> en 1000 ms: context deadline exceeded
```

Se despliega un binario más nuevo que el esquema: se simula quitando la migración 2.

```
$ curl /salud/listo
{"esperado":2,"esquema":1,"listo":false,"motivo":"esquema 1, el binario espera 2: falta migrar"}
  [503 en 0.003025 s]
[código de salida 0]
```

```
$ docker inspect -f '{{.State.Health.Status}}' g685-api
unhealthy
[código de salida 0]
```

(unhealthy a los 7 s: intervalo de 5 s y 3 reintentos)

```
$ api migrar
aplicada 00002_tareas.sql (6 ms)
esquema en la versión 2 (1 migraciones nuevas)
[código de salida 0]
```

```
$ curl /salud/listo
{"esperado":2,"esquema":2,"listo":true,"motivo":""}
  [200 en 0.001429 s]
[código de salida 0]
```

(healthy otra vez a los 4 s)

## Lección 11 — Apagado ordenado, medido

20 peticiones de 2 s en vuelo; a los 300 ms, SIGTERM. Tres servidores:

```
$ go run ./lecciones/11_apagado
  sin-senal  ->  0 de 20 OK  · el proceso sale a los    0 ms  · 20 con «EOF»
  ingenuo    ->  0 de 20 OK  · el proceso sale a los    0 ms  · 20 con «EOF»
  plantilla  -> 20 de 20 OK  · el proceso sale a los 2150 ms
[código de salida 0]
```

Y con la imagen de verdad y `docker stop`. Para tener peticiones en vuelo, una sesión de psql bloquea la tabla tareas 3 s; las 20 peticiones GET /tareas esperan al candado.

```
$ docker stop g685-api   (con las 20 peticiones esperando)
g685-api
(docker stop tardó 1.3 s)
[código de salida 0]
```

```
$ códigos HTTP que recibieron las 20 peticiones
200: 20
```

```
$ docker logs -f g685-api | grep -E 'apagando|apagado'
{"time":"2026-10-28T18:23:05.585499986Z","level":"INFO","msg":"apagando","drenaje":"0s","max":"8s"}
{"time":"2026-10-28T18:23:06.660857692Z","level":"INFO","msg":"apagado","ms":1075}
[código de salida 0]
```

```
$ docker wait g685-api   (código con el que salió)
0
[código de salida 0]
```

## Lección 12 — CI de GitHub Actions, corrido en local con act

```
$ act --version
act version 0.2.90
[código de salida 0]
```

```
$ act -l
Stage  Job ID  Job name  Workflow name  Workflow file  Events           
0      probar  probar    ci             ci.yml         push,pull_request
1      imagen  imagen    ci             ci.yml         push,pull_request
time="2026-10-28T18:23:11Z" level=info msg="Using docker host 'unix:///var/run/docker.sock', and daemon socket 'unix:///var/run/docker.sock'"
[código de salida 0]
```

```
$ act push -P ubuntu-latest=catthehacker/ubuntu:act-latest
[probar]   ✅  Success - Main actions/checkout@3d3c42e5aac5ba805825da76410c181273ba90b1 [33.988433ms]
[probar]   ✅  Success - Main usuarios de la base [3.711535333s]
[probar]   ✅  Success - Main go vet [6.102478456s]
[probar]    ok  	tareas/internal/config	1.010s
[probar]    ok  	tareas/internal/cuentas	3.993s
[probar]    ok  	tareas/internal/servidor	5.423s
[probar]    ok  	tareas/internal/tareas	6.870s
[probar]    ok  	tareas/internal/web	1.014s
[probar]   ✅  Success - Main tests (con -race y sin caché) [15.922079504s]
[probar] 🏁  Job succeeded
[imagen]   ✅  Success - Main actions/checkout@3d3c42e5aac5ba805825da76410c181273ba90b1 [28.901014ms]
[imagen]   ✅  Success - Main construir la imagen [749.813771ms]
[imagen]    17.5MB
[imagen]   ✅  Success - Main tamaño [119.255853ms]
[imagen] 🏁  Job succeeded
[act terminó con código 0]
(act tardó 62 s en total)
[código de salida 0]
```

```
$ docker image inspect catthehacker/ubuntu:act-latest (digest)
catthehacker/ubuntu@sha256:c58e2b364da03b0c804c7d660f2ecbedf2f221a382b9baa0b344b0144780ff43
[código de salida 0]
```

## Lección 13 — Usarla de plantilla

```
$ copiar, renombrar el módulo, compilar y probar (en un directorio de mktemp)
imports a cambiar: 13 archivos
quedan con el nombre viejo: 0
module github.com/tu-usuario/inventario
ok  	github.com/tu-usuario/inventario/internal/config	0.001s
ok  	github.com/tu-usuario/inventario/internal/cuentas	0.002s
ok  	github.com/tu-usuario/inventario/internal/servidor	0.002s
ok  	github.com/tu-usuario/inventario/internal/tareas	0.003s
ok  	github.com/tu-usuario/inventario/internal/web	0.003s
archivos .go: 17 de código, 6 de tests
líneas de Go: 1395 de código, 410 de tests
[código de salida 0]
```

docker compose de la plantilla, en una copia temporal y con los puertos quitados (para no chocar con nada de la máquina):

```
$ docker compose up -d --build --wait
(levantado en 10 s)
[código de salida 0]
```

```
$ docker compose ps -a
api  Up 5 seconds (healthy)
bd  Up 8 seconds (healthy)
migrar  Exited (0) 6 seconds ago
[código de salida 0]
```

```
$ docker compose logs migrar
aplicada 00001_usuarios.sql (6 ms)
aplicada 00002_tareas.sql (3 ms)
esquema en la versión 2 (2 migraciones nuevas)
[código de salida 0]
```

```
$ curl http://api:8080/salud/listo
{"esperado":2,"esquema":2,"listo":true,"motivo":""}
[código de salida 0]
```

## Resultado

Fallos inesperados: 0

