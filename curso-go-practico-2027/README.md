# Go práctico 2027 — Golandia (curso gratis, 5 partes)

Código de estos vídeos del canal Golandia:

- [Go práctico 2027 · Parte 1 de 5: API REST con net/http, JSON y códigos de estado](https://youtu.be/sk9-OLOIo18)
- [Go práctico 2027 · Parte 2 de 5: Postgres con pgx, migraciones y tiempos límite](https://youtu.be/RUqvBZynItM)
- [Go práctico 2027 · Parte 3 de 5: pruebas de tabla, httptest, Postgres real y -race](https://youtu.be/gOAtDBQgHK8)
- [Go práctico 2027 · Parte 4 de 5: configuración, slog, apagado ordenado y Docker de 15 MB](https://youtu.be/_WLInoYV0e0)
- [Go práctico 2027 · Parte 5 de 5: compose, health checks y la base de datos que se cae](https://youtu.be/8TV2zKd_BrE)

---

Continuación del «Curso de Go desde cero 2027». Un solo proyecto, una API de tareas, que crece parte a parte.
Cada carpeta `parteN/tareas/` es el proyecto completo tal como queda al terminar esa parte.

| parte | qué se añade |
|---|---|
| `parte1/` | API REST con `net/http` y los patrones de ruta de Go 1.22+, JSON, códigos de estado, validación. Almacén en memoria |
| `parte2/` | Postgres con pgx (`pgxpool`), migraciones con `go:embed`, tiempo límite por consulta, pool de conexiones, `/salud` |
| `parte3/` | Pruebas de tabla, `httptest`, integración con Postgres en un contenedor, cobertura y `-race` |
| `parte4/` | Configuración por variables de entorno, logs JSON con `slog`, apagado ordenado, Docker multi-stage (`scratch`) |
| `parte5/` | `compose.yaml` con la BD, health checks (`tareas salud`), la BD se cae y vuelve, reinicio automático |

    python ejecutar_todo.py          # las 5 partes en Docker; escribe parteN/SALIDA.md
    python ejecutar_todo.py 3        # solo la parte 3

Necesita Docker y Python 3.10+. Go va dentro de `golang:1.27.2`; Postgres es `postgres:18.6-alpine`.
Todo lo que crea lleva el prefijo `gl-curso-` y se borra por nombre al terminar
(`python ejecutar_todo.py --limpiar` borra también los volúmenes de caché de Go).

Sin Docker, con Go y un Postgres propios:

    cd parte4/tareas
    DATABASE_URL='postgres://usuario:clave@localhost:5432/tareas?sslmode=disable' go run .
