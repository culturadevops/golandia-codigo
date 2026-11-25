# Curso de Rust para crear CLI 2027 — Golandia (miembros)

Código de estos vídeos del canal Golandia:

- [Curso de Rust para crear CLI 2027 · Parte 1 de 3: clap, subcomandos y códigos de salida](https://youtu.be/p4DtHpsvvp8) (miembros)
- [Curso de Rust para crear CLI 2027 · Parte 2 de 3: errores, configuración, JSON y colores](https://youtu.be/fQQlc0JeULI) (miembros)
- [Curso de Rust para crear CLI 2027 · Parte 3 de 3: tests, release y binario estático](https://youtu.be/IqeUOk5Eqdg) (miembros)

> **El vídeo de este código está disponible para los miembros del canal: https://youtu.be/p4DtHpsvvp8**
> **El vídeo de este código está disponible para los miembros del canal: https://youtu.be/fQQlc0JeULI**
> **El vídeo de este código está disponible para los miembros del canal: https://youtu.be/IqeUOk5Eqdg**

---

A lo largo del curso se construye **fuga**, una CLI que busca secretos filtrados (claves de AWS,
tokens de GitHub, claves privadas, contraseñas) en archivos de texto, pensada para CI.

| carpeta | qué hay |
|---|---|
| `fuga/src/main.rs` | la herramienta completa (clap, anyhow, serde, toml, regex) con sus tests unitarios |
| `base/` | `Cargo.toml` con las versiones fijadas con `=` y su `Cargo.lock` |
| `datos/` | archivos de prueba; todos los secretos son FALSOS (la clave de AWS es la de ejemplo de su documentación) |
| `parte1/` | de `env::args` a clap, subcomandos, stdout/stderr, tubería rota y códigos de salida |
| `parte2/` | errores con anyhow, configuración por capas, `--json` para jq y colores solo en terminal |
| `parte3/` | tests unitarios, tests de la CLI con assert_cmd, tamaño y velocidad en release, binario estático musl |

Códigos de salida de `fuga`: `0` limpio · `1` hay secretos · `2` uso incorrecto (clap) · `3` no pudo trabajar.

    python ejecutar_todo.py            # todo en Docker y regenera SALIDA.md
    python ejecutar_todo.py parte3     # solo una parte, por pantalla
    python ejecutar_todo.py --limpiar  # además borra la imagen y los volúmenes de caché

Con Rust instalado: `cd base && cp ../fuga/src/main.rs src/ && cargo run -- escanear ../datos/*`.
