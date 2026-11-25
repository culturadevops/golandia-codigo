# Banco de «¿Go o Rust? La revancha» (golandia 958)

Código del vídeo [«¿Go o Rust? La revancha: LTO, musl, mimalloc, actix y el modo debug (lo medí)»](https://youtu.be/kGKGZ4n5SCw) del canal Golandia.

Mismo contrato que toda la serie «¿Go o X? Lo medí» (/suma y /trabajo, 2 CPU, 1 GiB, `hey -c16`), ahora con las variantes de Rust que faltaron en el vídeo original: repartidor de memoria (mimalloc, jemalloc), musl, perfil con LTO fat + codegen-units=1 + panic=abort, actix-web y el modo debug; y Go con GOGC=400.

| archivo | qué es |
|---|---|
| `pruebas_go_rust_revancha.py` | el banco (usa `../banco-go-o-x/`) |
| `axum/`, `actix/` | los servidores de Rust |
| `Dockerfile.*` | una imagen por variante |
| `SALIDA-corrida-1.md`, `SALIDA-corrida-3.md` | resultados (la corrida 2 se cortó por un reinicio y no cuenta) |

Ejecutar (necesita Docker): `python pruebas_go_rust_revancha.py`
