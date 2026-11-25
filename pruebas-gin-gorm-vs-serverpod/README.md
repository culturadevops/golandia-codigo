# ¿Gin + GORM o Serverpod? — banco del vídeo 866 (golandia)

Código del vídeo [«¿Gin + GORM o Serverpod? Lo medí: la misma API con Postgres en Go y en Dart»](https://youtu.be/fsoKeGqW-I0) del canal Golandia.

---

La misma API (leer una tarea por id, 20 autores con sus tareas con Preload/include y con N+1 a mano, crear una tarea)
en Go con Gin + GORM (`go/`) y en Dart con Serverpod 4.1.2 (`gs_api/`), contra el mismo Postgres 18.6 y las mismas
tablas (las crea la migración de Serverpod), con 1.000 autores × 10 tareas resembrados antes de cada servidor.

    python banco.py          # → SALIDA.md (toma él mismo el candado <proyecto>/_logs/banco.lock)
    python banco_extra.py    # → SALIDA-extra.md (conexiones nuevas y CPU) — NO llegó a correr el 2026-11-16 (disco lleno)

- `Dockerfile.go`: golang:1.27.2 → scratch. `Dockerfile.serverpod`: el que genera `serverpod create`, con alpine fijado a 3.24.4.
- `_sp.sh`: corre el CLI de serverpod en `dart:3.13.7` con un `flutter` de mentira (el CLI exige Flutter en el PATH aunque el
  proyecto sea solo de servidor; lo único que le pide es la versión: ver `_stub_flutter.log`).
- Go: `POOL=n` fija `SetMaxOpenConns(n)` y `SetMaxIdleConns(n)`; sin `POOL`, los valores por defecto de database/sql.
- Docker: red `gs-red` (10.251.77.0/24), contenedores `gs-pg` y `gs-srv`, imágenes `gs-go:1`, `gs-serverpod:1`,
  `gs-flutter:3.44.0` (= ghcr.io/cirruslabs/flutter:stable, 7 GB, ya no hace falta), volúmenes `gs-pubcache`, `gs-pubcache-fl`, `gs-gomod`.
