# Curso de Rust para DevOps 2027 — Golandia (miembros)

Código de estos vídeos del canal Golandia:

- [Curso de Rust para DevOps 2027 · Parte 1 de 3: async, chequeos HTTP y reintentos](https://youtu.be/zSzDk6rf9KQ) (miembros)
- [Curso de Rust para DevOps 2027 · Parte 2 de 3: configuración YAML/TOML y comandos](https://youtu.be/R7eLE3ZCeeY) (miembros)
- [Curso de Rust para DevOps 2027 · Parte 3 de 3: métricas, apagado limpio e imagen mínima](https://youtu.be/lYaZXUgp-nA) (miembros)

> **El vídeo de este código está disponible para los miembros del canal: https://youtu.be/zSzDk6rf9KQ**
> **El vídeo de este código está disponible para los miembros del canal: https://youtu.be/R7eLE3ZCeeY**
> **El vídeo de este código está disponible para los miembros del canal: https://youtu.be/lYaZXUgp-nA**

---

Sigue al Curso básico de Rust 2027. Todo ejecutable en Docker; cada número del vídeo sale de `SALIDA.md`.

| parte | tema |
|---|---|
| `parte1/` | async con tokio (tareas contra hilos, el sleep que bloquea), chequeos HTTP concurrentes con reqwest, timeouts, el 500 que parece «arriba», reintentos con backoff |
| `parte2/` | configuración YAML y TOML con serde, el campo mal escrito que se ignora, validación, ejecutar comandos externos y el timeout que deja procesos huérfanos |
| `parte3/` | servicio con axum: /health, /metrics para Prometheus, logs JSON con tracing, apagado limpio con SIGTERM, imagen glibc contra musl sobre scratch, y `docker stop` con PID 1 |

Binarios en `src/bin/`: `objetivo` (servidor de pruebas), `tareas`, `salud`, `config`, `comandos`, `servicio`.

    python ejecutar_todo.py          # todo y regenera SALIDA.md (construye y borra las imágenes rdevops-servicio:*)
    python ejecutar_todo.py parte1   # solo una parte, por pantalla

La caché de crates queda en el volumen `rdevops-registry`: `docker volume rm rdevops-registry` para borrarla.
