# Curso básico de Rust 2027 — Golandia (miembros)

Código de estos vídeos del canal Golandia:

- [Curso básico de Rust 2027 · Parte 1 de 5: Cargo, variables, tipos y funciones](https://youtu.be/mC9aXrhLKjU) (miembros)
- [Curso básico de Rust 2027 · Parte 2 de 5: if, bucles y ownership](https://youtu.be/08ownNxTC14) (miembros)
- [Curso básico de Rust 2027 · Parte 3 de 5: préstamos, referencias y slices](https://youtu.be/yxUOrTpsO2A) (miembros)
- [Curso básico de Rust 2027 · Parte 4 de 5: structs, enums, Option y match](https://youtu.be/2XqOBWur_IE) (miembros)
- [Curso básico de Rust 2027 · Parte 5 de 5: Vec, HashMap, Result y tu primer programa](https://youtu.be/Hiv3hLDIY_w) (miembros)
- [Curso básico de Rust 2027: 5 partes desde cero, ya para los miembros de Golandia](https://youtu.be/kt6Mf0N5Dtw)

> **El vídeo de este código está disponible para los miembros del canal: https://youtu.be/mC9aXrhLKjU**
> **El vídeo de este código está disponible para los miembros del canal: https://youtu.be/08ownNxTC14**
> **El vídeo de este código está disponible para los miembros del canal: https://youtu.be/yxUOrTpsO2A**
> **El vídeo de este código está disponible para los miembros del canal: https://youtu.be/2XqOBWur_IE**
> **El vídeo de este código está disponible para los miembros del canal: https://youtu.be/Hiv3hLDIY_w**

---

Todo el código de las 5 partes del curso, ejecutable. Cada lección es una carpeta con sus `.rs`
y un `demo.sh` que crea un proyecto de Cargo de verdad, compila y ejecuta.

| parte | tema |
|---|---|
| `parte1/` | Cargo, variables, mutabilidad, sombreado, tipos, desbordamiento, funciones, tuplas y arrays |
| `parte2/` | if como expresión, bucles, ownership: mover, clonar, Copy, funciones y Drop |
| `parte3/` | Texto UTF-8, referencias, `&mut`, la regla del préstamo, referencias colgantes y slices |
| `parte4/` | Structs, Debug, impl y métodos, enums, Option y match exhaustivo |
| `parte5/` | Vec, HashMap, panic, Result, el operador `?` y el proyecto final `palabras` |

Los errores de compilación son a propósito: cada lección enseña lo que dice el compilador.

    python ejecutar_todo.py          # todo en Docker (rust:1.99.1) y regenera SALIDA.md
    python ejecutar_todo.py parte3   # solo una parte, por pantalla

Sin Docker, con Rust instalado (rustup.rs): copia cualquier `.rs` a `src/main.rs` de un
`cargo new` y `cargo run`.
