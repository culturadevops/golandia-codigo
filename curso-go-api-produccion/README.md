# Una API en Go de cero a producción

Código del vídeo [«API en Go de cero a producción: Postgres, tests, Docker y CI en una plantilla (curso completo)»](https://youtu.be/4e2UQwHGZss) del canal Golandia.

> **El vídeo de este código está disponible para los miembros del canal: https://youtu.be/4e2UQwHGZss**

---

Repositorio del curso para miembros de **Golandia**. Es dos cosas a la vez:

1. **Una plantilla**: una API de tareas pequeña pero completa, lista para copiar y convertir en
   tu servicio. Router de `net/http`, middlewares, configuración validada, Postgres con
   migraciones y un usuario sin permisos de DDL, sesiones con bcrypt, errores JSON coherentes,
   tests de integración contra Postgres real, imagen de 17,5 MB, healthcheck y apagado ordenado,
   y un CI de GitHub Actions.
2. **Las 13 lecciones del vídeo**, en `lecciones/`, cada una con la trampa que no da error y su
   medida. `ejecutar_todo.py` las corre todas y escribe la salida literal en `SALIDA.md`.

## Requisitos

- **Docker** (probado con Docker 29.6.2). Go, Postgres, curl y act corren en contenedores.
- **Python 3.10+** solo para `ejecutar_todo.py`.
- Go 1.26+ si quieres compilar o probar fuera de Docker (el curso usa `golang:1.27.2`).

## Arrancarla en local

```sh
docker compose up --build
curl localhost:8080/salud/listo
```

Compose levanta Postgres 18.6 (crea los usuarios con `sql/roles.sql`), aplica las migraciones
con el usuario `migrador` como paso aparte y arranca la API con el usuario `app`.

```sh
curl -X POST localhost:8080/usuarios -d '{"email":"ana@ejemplo.com","clave":"una clave larga"}'
curl -X POST localhost:8080/sesiones -d '{"email":"ana@ejemplo.com","clave":"una clave larga"}'
# -> {"token":"…","vence":"…"}
curl -X POST localhost:8080/tareas -H "Authorization: Bearer <token>" -d '{"titulo":"probar la API"}'
curl localhost:8080/tareas -H "Authorization: Bearer <token>"
```

## La API

| Método y ruta | Qué hace |
|---|---|
| `POST /usuarios` | registra `{email, clave}` (clave de 10 a 72 bytes) |
| `POST /sesiones` | inicia sesión: devuelve `{token, vence}` |
| `DELETE /sesiones` | cierra la sesión: el token deja de valer al instante |
| `GET /tareas?hecha=&limite=&despues=` | lista, paginada por id (`siguiente`) |
| `POST /tareas` | crea `{titulo, prioridad}` (prioridad 1-3, por defecto 2) |
| `GET/PATCH/DELETE /tareas/{id}` | una tarea; la de otro usuario es 404 |
| `GET /salud/vivo` | liveness: el proceso responde (no mira la base) |
| `GET /salud/listo` | readiness: base con plazo de 1 s, esquema al día y no apagándose |

Todos los errores tienen la misma forma:

```json
{"error":{"codigo":"validacion","mensaje":"hay campos con problemas",
          "campos":{"titulo":"obligatorio","prioridad":"1, 2 o 3"}},"request_id":"…"}
```

## Configuración (variables de entorno)

| Variable | Por defecto | |
|---|---|---|
| `PUERTO` | — (obligatoria) | 1-65535 |
| `BD_URL` | — (obligatoria) | `postgres://app:…@host:5432/tareas`: el usuario **sin** DDL |
| `BD_URL_MIGRAR` | — | solo para `api migrar`: el usuario dueño del esquema |
| `TIEMPO_PETICION` | `10s` | plazo de cada petición |
| `APAGADO_MAX` | `8s` | lo que espera a las peticiones en vuelo; menor que el plazo de tu orquestador (Docker: 10 s; Kubernetes: 30 s) |
| `DRENAJE` | `0s` | tiempo con `listo`=503 antes de cerrar, para que el balanceador se entere |
| `NIVEL_LOG` | `info` | debug, info, warn, error |
| `COSTO_BCRYPT` | `12` | 10-14 |

Si algo falta o está mal, el proceso no arranca y dice **todos** los problemas a la vez.

## Estructura

```
cmd/api/            main.go (solo señales y os.Exit) y las órdenes: servir, migrar, salud
internal/config/    entorno -> Config, validado; DSN que no imprime la clave
internal/web/       errores JSON, lectura estricta del cuerpo, middlewares
internal/cuentas/   usuarios, bcrypt, sesiones en Postgres
internal/tareas/    el dominio: modelo, validación, almacén Postgres, handlers
internal/servidor/  rutas + orden de middlewares, salud, http.Server y apagado
internal/bd/        pool, migraciones (goose embebido), versión del esquema
internal/pruebabd/  Postgres real para los tests de integración
migraciones/        los .sql, metidos en el binario con go:embed
sql/roles.sql       los dos usuarios (lo corre el superusuario una vez)
Dockerfile          multietapa a scratch, usuario 65532, HEALTHCHECK con el propio binario
.github/workflows/  vet + tests con -race contra Postgres + imagen
lecciones/          las demos del curso (bórralo al usarla de plantilla)
```

## Tests

```sh
go test ./...                       # unitarios; los de integración se SALTAN sin base
CI=true go test ./...               # en CI, sin base, FALLAN (así no hay verde falso)

# con la base de tests (docker compose up bd):
export TEST_BD_URL=postgres://app:app@localhost:5432/tareas_test?sslmode=disable
export TEST_BD_URL_MIGRAR=postgres://migrador:migrador@localhost:5432/tareas_test?sslmode=disable
go test -race -count=1 ./...        # -count=1: la caché de go test no sabe que la base cambió
```

## Usarla de plantilla

```sh
cp -r curso-go-api-produccion mi-servicio && cd mi-servicio
rm -rf .git lecciones ejecutar_todo.py SALIDA.md
go mod edit -module github.com/tu-usuario/mi-servicio
grep -rl '"tareas/' --include=*.go . | xargs sed -i 's#"tareas/#"github.com/tu-usuario/mi-servicio/#'
go build ./... && go test ./...
```

Luego cambia `internal/tareas` por tu dominio y añade sus migraciones en `migraciones/`
(nunca edites una que ya se aplicó: lección 5).

## Las lecciones

| # | Lección | Cómo correrla |
|---|---|---|
| 1 | La estructura: main solo llama a run() | `go run ./lecciones/01_estructura -modo fatal` |
| 2 | El router de net/http (Go 1.22+) | `go run ./lecciones/02_router` |
| 3 | Middlewares: request id, slog, recover, plazo | `go run ./lecciones/03_middlewares` |
| 4 | Configuración validada al arrancar | `go run ./lecciones/04_config` |
| 5 | Migraciones y un usuario sin DDL | `DEMO_BD_URL=… go run ./lecciones/05_migraciones` |
| 6 | Autenticación: sesiones y bcrypt | `APP_BD_URL=… go run ./lecciones/06_auth` |
| 7 | Validación y errores JSON | `APP_BD_URL=… go run ./lecciones/07_errores` |
| 8 | Tests de integración contra Postgres | `go test ./...` (ver arriba) |
| 9 | Dockerfile multietapa, medido | `docker build -f lecciones/09_docker/Dockerfile.ingenuo .` |
| 10 | Healthcheck: vivo y listo | `APP_BD_URL=… go run ./lecciones/10_salud` |
| 11 | Apagado ordenado, medido | `go run ./lecciones/11_apagado` (Linux o Docker) |
| 12 | CI de GitHub Actions con act | ver `ejecutar_todo.py` (act v0.2.90 en un contenedor) |
| 13 | Usarla de plantilla | arriba |

Lo más cómodo es no correrlas a mano: `python ejecutar_todo.py` levanta todo en Docker (con
nombres `g685-*`), corre las 13 y regenera `SALIDA.md`. Tarda unos minutos, sobre todo por act.
Al terminar, aunque una lección falle, borra todo lo que creó: contenedores (todos con `--rm`),
red, volúmenes, las imágenes `g685-tareas:*` y `tareas:ci`, y la imagen del runner de act.
`python ejecutar_todo.py --limpiar` hace solo esa limpieza.

## Versiones

Go 1.27.2 (`golang:1.27.2`) · PostgreSQL 18.6 (`postgres:18.6-alpine`) · pgx v5.11.0 ·
goose v3.29.0 · golang.org/x/crypto v0.58.0 · act v0.2.90 con `catthehacker/ubuntu:act-latest`.
Las cifras de `SALIDA.md` dependen de la máquina (Docker Desktop sobre Windows).

Sin ORM y sin framework web: solo la biblioteca estándar, pgx, goose y bcrypt.
