# SALIDA — ¿Gin + GORM o Serverpod? (golandia 866)

> Generado por `banco.py` el 2026-11-16 21:49:06Z. **No se escribe a mano.** Servidores con 2 CPU y 1 GiB; Postgres 18.6 con 2 CPU y 1 GiB; carga con `hey -c 16` y 4 CPU desde otro contenedor. Datos: 1.000 autores × 10 tareas, resembrados antes de cada servidor.

## Versiones

```
go version go1.27.2 linux/amd64
github.com/gin-gonic/gin v1.12.0
gorm.io/gorm v1.31.2
gorm.io/driver/postgres v1.6.3
github.com/jackc/pgx/v5 v5.10.0
serverpod 4.1.2
dart (imagen de build de Serverpod) Dart SDK version: 3.12.2 (stable) (Tue Jun 9 01:11:39 2026 -0700) on "linux_x64"
PostgreSQL 18.6 (Debian 18.6-1.pgdg13+2) on x86_64-pc-linux-gnu, compiled by gcc (Debian 14.2.0-19) 14.2.0, 64-bit
```

## 0 · ¿Contestan lo mismo? (antes de medir)

- leer /tarea 4242 → Gin+GORM {"id": 4242, "titulo": "tarea-425-2", "hecha": false, "autorId": 425}
                    → Serverpod {"id": 4242, "titulo": "tarea-425-2", "hecha": false, "autorId": 425}
- leer una tarea por id: IGUAL (JSON de 69 bytes sin __className__)
- 20 autores con sus tareas (Preload / include): IGUAL (JSON de 14174 bytes sin __className__)
- 20 autores con sus tareas, N+1 a mano: IGUAL (JSON de 14174 bytes sin __className__)
- crear: Gin+GORM (200, 10001, '10001', 'carga false 7') · Serverpod (200, 10001, '10001', 'carga false 7')

Serverpod añade `"__className__"` a cada objeto del JSON. Sin ese campo, las respuestas son idénticas.

## 1 · Lo que escribes tú y lo que se genera

| | archivos | líneas |
|---|---|---|
| Gin + GORM, escrito a mano (go/main.go) | 1 | 100 |
| Serverpod, escrito a mano (2 modelos .spy.yaml + el endpoint) | 3 | 47 |
| Serverpod, generado en el servidor (lib/src/generated) | 6 | 2019 |
| Serverpod, generado: el cliente de Dart (gs_api_client/lib) | 5 | 528 |

El proyecto de Serverpod entero (lo crea `serverpod create --template server`, sin .dart_tool): 55 archivos.

`serverpod create` y `serverpod generate` (CLI 4.1.2) se niegan a correr sin `flutter` en el PATH aunque el proyecto sea solo de servidor: «ERROR: Failed to run serverpod. You need to have flutter installed and in your $PATH». Con un `flutter` de mentira que solo apunta con qué lo llaman, esto es todo lo que le pidió:

```
STUB flutter: --version  (×3)
STUB flutter: --version --machine  (×4)
```

Tablas que deja la migración de Serverpod: 17 (2 mías: autor y tarea; el resto, serverpod_*).

Dependencias: Go `go list -m all` → 65 módulos (sin contar el propio). Dart pubspec.lock del workspace → 104 paquetes.

## 2 · Construir la imagen desde cero (`docker build --no-cache`, bases ya descargadas)

| | tiempo | imagen (comprimida) | binario |
|---|---|---|---|
| Gin + GORM | 20.6 s | 9.1 MiB | 25.5 MiB |
| Serverpod | 20.3 s | 11.5 MiB | 12.7 MiB |

Gin + GORM: golang:1.27.2 → scratch (CGO_ENABLED=0, -s -w). Serverpod: el Dockerfile que genera `serverpod create` (dart:3.12.2 → alpine:3.24.4, `dart build cli`). Los dos tiempos incluyen bajar las dependencias.

## 3 · Las consultas que manda cada ORM (log_statement=all en Postgres)

### Gin + GORM (pool por defecto)

```
leer una tarea por id: 1 consulta(s)
   execute stmtcache_44d43ad6cb61e44f7eb6a70e36576cd3f0ee928b74b68c0c: SELECT * FROM "tarea" WHERE "tarea"."id" = $1 ORDER BY "tarea"."id" LIMIT $2
20 autores con sus tareas (Preload / include): 2 consulta(s)
   execute stmtcache_c36e14641d0e2efde3729df182dd7265f1caa07c214c6136: SELECT * FROM "autor" ORDER BY id LIMIT $1
   execute stmtcache_50580116ab6469ac37d27dc9638d0e03d094f3484c3fbe89: SELECT * FROM "tarea" WHERE "tarea"."autorId" IN ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16,$17,$18,$19,$20) ORDER BY id
20 autores con sus tareas, N+1 a mano: 21 consulta(s)
   execute stmtcache_c36e14641d0e2efde3729df182dd7265f1caa07c214c6136: SELECT * FROM "autor" ORDER BY id LIMIT $1
   execute stmtcache_524b1ae95ec9eac0215219ffe2b19d41bb3caff864ae3343: SELECT * FROM "tarea" WHERE "autorId" = $1 ORDER BY id
   execute stmtcache_524b1ae95ec9eac0215219ffe2b19d41bb3caff864ae3343: SELECT * FROM "tarea" WHERE "autorId" = $1 ORDER BY id
   ... y 18 más iguales
crear una tarea (INSERT): 3 consulta(s)
   statement: begin
   execute stmtcache_fc88826c411e260bd4f47bb09352e7dbfcbc911e0a746b73: INSERT INTO "tarea" ("titulo","hecha","autorId") VALUES ($1,$2,$3) RETURNING "id"
   statement: commit
65 s en reposo, sin ninguna petición: 0 consulta(s)
```

### Serverpod

```
leer una tarea por id: 1 consulta(s)
   execute <unnamed>/p/25: SELECT "tarea"."id" AS "tarea.id", "tarea"."titulo" AS "tarea.titulo", "tarea"."hecha" AS "tarea.hecha", "tarea"."autorId" AS "tarea.autorId" FROM "tarea" WHERE "tarea"."id" = 4242 LIMIT 1
20 autores con sus tareas (Preload / include): 2 consulta(s)
   execute <unnamed>/p/26: SELECT "autor"."id" AS "autor.id", "autor"."nombre" AS "autor.nombre" FROM "autor" ORDER BY "autor"."id" ASC NULLS LAST LIMIT 20
   execute <unnamed>/p/27: SELECT "tarea"."id" AS "tarea.id", "tarea"."titulo" AS "tarea.titulo", "tarea"."hecha" AS "tarea.hecha", "tarea"."autorId" AS "tarea.autorId" FROM "tarea" WHERE "tarea"."autorId" IN (1, 2, 3, 4, 5, 6, 7,...
20 autores con sus tareas, N+1 a mano: 21 consulta(s)
   execute <unnamed>/p/28: SELECT "autor"."id" AS "autor.id", "autor"."nombre" AS "autor.nombre" FROM "autor" ORDER BY "autor"."id" ASC NULLS LAST LIMIT 20
   execute <unnamed>/p/29: SELECT "tarea"."id" AS "tarea.id", "tarea"."titulo" AS "tarea.titulo", "tarea"."hecha" AS "tarea.hecha", "tarea"."autorId" AS "tarea.autorId" FROM "tarea" WHERE "tarea"."autorId" = 1 ORDER BY "tarea"."id...
   execute <unnamed>/p/30: SELECT "tarea"."id" AS "tarea.id", "tarea"."titulo" AS "tarea.titulo", "tarea"."hecha" AS "tarea.hecha", "tarea"."autorId" AS "tarea.autorId" FROM "tarea" WHERE "tarea"."autorId" = 2 ORDER BY "tarea"."id...
   ... y 18 más iguales
crear una tarea (INSERT): 1 consulta(s)
   execute <unnamed>/p/49: INSERT INTO "tarea" ("titulo", "hecha", "autorId") VALUES ('carga', FALSE, 7) RETURNING *
65 s en reposo, sin ninguna petición: 14 consulta(s) en serverpod_health_connection_info, serverpod_health_metric, serverpod_runtime_settings, serverpod_session_log
```

## 4 · Arranque hasta el primer 200 (3 arranques, Postgres ya listo)

| | arranques (ms) | mediana | memoria en reposo |
|---|---|---|---|
| Gin + GORM (pool por defecto) | 199 · 178 · 192 | 192 ms | 5.5 MiB |
| Serverpod | 358 · 392 · 356 | 358 ms | 20.2 MiB |

## 5 · Carga, corrida 1 (`hey -c 16`, 5 s de calentamiento y 15 s medidos)

| variante | escena | pet/s | p50 | p99 | memoria bajo carga | códigos |
|---|---|---|---|---|---|---|
| Gin + GORM (pool por defecto) | leer una tarea por id | 1.998 | 0.9 ms | 82.1 ms | 14.8 MiB | 200×30006 |
| Gin + GORM (pool por defecto) | 20 autores con sus tareas (Preload / include) | 1.129 | 3.2 ms | 77.5 ms | 18.2 MiB | 200×16946 |
| Gin + GORM (pool por defecto) | 20 autores con sus tareas. N+1 a mano | 260 | 89.7 ms | 101.1 ms | 20.7 MiB | 200×3920 |
| Gin + GORM (pool por defecto) | crear una tarea (INSERT) | 1.200 | 3.0 ms | 83.9 ms | 19.3 MiB | 200×18042 |
| Gin + GORM (POOL=10) | leer una tarea por id | 61.560 | 0.2 ms | 0.8 ms | 13.7 MiB | 200×923430 |
| Gin + GORM (POOL=10) | 20 autores con sus tareas (Preload / include) | 5.372 | 2.3 ms | 25.7 ms | 13.8 MiB | 200×80596 |
| Gin + GORM (POOL=10) | 20 autores con sus tareas. N+1 a mano | 428 | 11.0 ms | 84.9 ms | 13.8 MiB | 200×6432 |
| Gin + GORM (POOL=10) | crear una tarea (INSERT) | 8.437 | 1.8 ms | 3.1 ms | 13.4 MiB | 200×126576 |
| Serverpod | leer una tarea por id | 9.088 | 1.7 ms | 3.0 ms | 22.3 MiB | 200×136342 |
| Serverpod | 20 autores con sus tareas (Preload / include) | 1.430 | 11.2 ms | 13.8 ms | 71.9 MiB | 200×21460 |
| Serverpod | 20 autores con sus tareas. N+1 a mano | 415 | 30.5 ms | 58.6 ms | 48.3 MiB | 200×6238 |
| Serverpod | crear una tarea (INSERT) | 6.151 | 2.5 ms | 5.4 ms | 48.6 MiB | 200×92280 |

## 6 · Carga, corrida 2 (`hey -c 16`, 5 s de calentamiento y 15 s medidos)

| variante | escena | pet/s | p50 | p99 | memoria bajo carga | códigos |
|---|---|---|---|---|---|---|
| Gin + GORM (pool por defecto) | leer una tarea por id | 2.024 | 0.9 ms | 81.8 ms | 15.5 MiB | 200×30405 |
| Gin + GORM (pool por defecto) | 20 autores con sus tareas (Preload / include) | 1.141 | 3.3 ms | 76.8 ms | 18.5 MiB | 200×17183 |
| Gin + GORM (pool por defecto) | 20 autores con sus tareas. N+1 a mano | 277 | 88.1 ms | 100.3 ms | 20.2 MiB | 200×4173 |
| Gin + GORM (pool por defecto) | crear una tarea (INSERT) | 1.148 | 3.5 ms | 80.0 ms | 19.2 MiB | 200×17248 |
| Gin + GORM (POOL=10) | leer una tarea por id | 61.394 | 0.2 ms | 0.8 ms | 13.8 MiB | 200×920979 |
| Gin + GORM (POOL=10) | 20 autores con sus tareas (Preload / include) | 5.772 | 2.2 ms | 22.5 ms | 13.8 MiB | 200×86626 |
| Gin + GORM (POOL=10) | 20 autores con sus tareas. N+1 a mano | 435 | 10.8 ms | 84.9 ms | 13.5 MiB | 200×6527 |
| Gin + GORM (POOL=10) | crear una tarea (INSERT) | 7.864 | 1.9 ms | 3.9 ms | 13.7 MiB | 200×117991 |
| Serverpod | leer una tarea por id | 9.287 | 1.7 ms | 2.8 ms | 21.9 MiB | 200×139323 |
| Serverpod | 20 autores con sus tareas (Preload / include) | 1.460 | 10.8 ms | 14.4 ms | 70.4 MiB | 200×21910 |
| Serverpod | 20 autores con sus tareas. N+1 a mano | 416 | 29.1 ms | 60.5 ms | 48.4 MiB | 200×6244 |
| Serverpod | crear una tarea (INSERT) | 6.180 | 2.5 ms | 4.0 ms | 47.5 MiB | 200×92715 |

