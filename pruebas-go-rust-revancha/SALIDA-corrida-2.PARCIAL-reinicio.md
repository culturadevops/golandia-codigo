# SALIDA — ¿Go o Rust? La revancha (golandia 958), corrida 2, con el banco de la serie

> Corrida del 2026-11-24 21:43, con el candado `_logs/banco.lock` tomado. Regenerable con `python investigacion/pruebas_go_rust_revancha.py 2` (mismo banco que 879, 865 y 902-905; el 698 es `recursos/pruebas-go-vs-rust/SALIDA.md`). ⛔ Ninguna cifra de blog: todo salió de aquí.

- go: `go version go1.27.2 linux/amd64` (net/http, solo librería estándar)
- rust: `rustc 1.99.1 (58cf8f9c3 2026-10-13) · cargo 1.99.1 (5f94df478 2026-08-27)` · alpine `host: x86_64-unknown-linux-musl`
- crates (Cargo.lock): axum 0.8.9 · tokio 1.54.1 · mimalloc 0.1.52 · tikv-jemallocator 0.6.1 · actix-web 4.20.0 (su tokio 1.54.1)
- Docker `29.6.2`. Servidores con 2 CPU y 1g; `hey -c 16` con 4 CPU desde otro contenedor.

## 1 · Construir la imagen desde cero (`docker build --no-cache`, dependencias incluidas)

| variante | qué se cronometra | segundos |
|---|---|---|
| go | go build | 6.3 |
| rust (698) | cargo build --release (axum, glibc) | 12.1 |
| rust tope | cargo build --profile tope (LTO fat, codegen-units=1, panic=abort) | 16.8 |
| rust + mimalloc | cargo build --release --features mimalloc (glibc) | 12.1 |
| rust musl (698) | cargo build --release en rust:1.99.1-alpine | 21.6 |
| rust musl + mimalloc | lo mismo con --features mimalloc | 24.6 |
| rust musl + jemalloc | apk add make + --features jemalloc (compila jemalloc) | 28.0 |
| actix-web | cargo build --release (actix-web 4, glibc) | 20.4 |
| rust debug | cargo build (sin --release) | 13.7 |

## 2 · Lo que despliegas

| imagen | base | tamaño |
|---|---|---|
| go | scratch + binario estático | 2.4 MiB |
| rust | debian:trixie-slim + binario release (el del 698) | 29.0 MiB |
| rust-tope | debian:trixie-slim + binario con LTO fat + codegen-units=1 + panic=abort | 28.9 MiB |
| rust-mimalloc | debian:trixie-slim + binario release con mimalloc | 29.1 MiB |
| rust-musl | scratch + binario estático musl (el del 698) | 0.6 MiB |
| rust-musl-mimalloc | scratch + binario estático musl con mimalloc | 0.7 MiB |
| rust-musl-jemalloc | scratch + binario estático musl con jemalloc | 0.8 MiB |
| actix | debian:trixie-slim + binario release de actix-web | 29.2 MiB |
| rust-debug | debian:trixie-slim + binario debug (sin --release) | 35.6 MiB |

| binario (lo que se copia a la imagen) | bytes | MiB |
|---|---|---|
| go (`-ldflags='-s -w'`) | 5873824 | 5.60 |
| rust | 1282944 | 1.22 |
| rust-tope | 863368 | 0.82 |
| rust-mimalloc | 1429096 | 1.36 |
| rust-musl | 1397832 | 1.33 |
| rust-musl-mimalloc | 1551496 | 1.48 |
| rust-musl-jemalloc | 1824456 | 1.74 |
| actix | 1836432 | 1.75 |
| rust-debug | 26502344 | 25.27 |

- CPU de Windows (Win32_Processor.LoadPercentage) antes de medir: 9 %
## Estado de la máquina antes de medir

- /proc/loadavg (desde un contenedor alpine, VM de Docker con 24 hilos): `2.55 2.73 2.09 3/513 6`
- contenedores ajenos en marcha: ninguno

## 0 · ¿Hacen lo mismo? (antes de medir)

- go: total 9990000, primeros ids [0, 1000, 2000, 3000, 4000]…, suma 99863963591 → igual que Go
- go, GOGC=400: total 9990000, primeros ids [0, 1000, 2000, 3000, 4000]…, suma 99863963591 → igual que Go
- rust (698): total 9990000, primeros ids [0, 1000, 2000, 3000, 4000]…, suma 99863963591 → igual que Go
- rust tope (LTO): total 9990000, primeros ids [0, 1000, 2000, 3000, 4000]…, suma 99863963591 → igual que Go
- rust + mimalloc: total 9990000, primeros ids [0, 1000, 2000, 3000, 4000]…, suma 99863963591 → igual que Go
- rust musl (698): total 9990000, primeros ids [0, 1000, 2000, 3000, 4000]…, suma 99863963591 → igual que Go
- rust musl + mimalloc: total 9990000, primeros ids [0, 1000, 2000, 3000, 4000]…, suma 99863963591 → igual que Go
- rust musl + jemalloc: total 9990000, primeros ids [0, 1000, 2000, 3000, 4000]…, suma 99863963591 → igual que Go
- actix-web: total 9990000, primeros ids [0, 1000, 2000, 3000, 4000]…, suma 99863963591 → igual que Go
- rust debug: total 9990000, primeros ids [0, 1000, 2000, 3000, 4000]…, suma 99863963591 → igual que Go

## 3 · Arranque: contenedor nuevo → primera respuesta 200 (mediana de 8) y memoria en reposo

| variante | 1ª respuesta (docker incluido) | «listo» (su propio reloj) | memoria en reposo |
|---|---|---|---|
| go | 204 ms | 0.0 ms | 2.7 MiB |
| go, GOGC=400 | 209 ms | 0.0 ms | 2.8 MiB |
| rust (698) | 217 ms | 0.0 ms | 1.9 MiB |
| rust tope (LTO) | 197 ms | 0.0 ms | 1.9 MiB |
| rust + mimalloc | 200 ms | 0.0 ms | 8.0 MiB |
| rust musl (698) | 204 ms | 0.1 ms | 2.0 MiB |
| rust musl + mimalloc | 193 ms | 0.1 ms | 7.7 MiB |
| rust musl + jemalloc | 203 ms | 0.1 ms | 2.1 MiB |
| actix-web | 203 ms | 0.1 ms | 2.3 MiB |
| rust debug | 190 ms | 0.1 ms | 2.0 MiB |

## 3b · Lo que dice cada servidor al arrancar (con --cpus 2)

- go → `LISTO_EN_MS 0.056`
- go, GOGC=400 → `LISTO_EN_MS 0.038`
- rust (698) → `LISTO_EN_MS 0.050 rust axum hilos=2 repartidor=sistema perfil=optimizado`
- rust tope (LTO) → `LISTO_EN_MS 0.063 rust axum hilos=2 repartidor=sistema perfil=optimizado`
- rust + mimalloc → `LISTO_EN_MS 0.048 rust axum hilos=2 repartidor=mimalloc perfil=optimizado`
- rust musl (698) → `LISTO_EN_MS 0.056 rust axum hilos=2 repartidor=sistema perfil=optimizado`
- rust musl + mimalloc → `LISTO_EN_MS 0.071 rust axum hilos=2 repartidor=mimalloc perfil=optimizado`
- rust musl + jemalloc → `LISTO_EN_MS 0.060 rust axum hilos=2 repartidor=jemalloc perfil=optimizado`
- actix-web → `LISTO_EN_MS 0.092 rust actix hilos=2`
- rust debug → `LISTO_EN_MS 0.137 rust axum hilos=2 repartidor=sistema perfil=debug`

## 4 · Carga con memoria: /trabajo (objetos, mapa, orden, JSON), a tope

`hey -c 16` desde otro contenedor; servidor con 2 CPU y 1g. «En frío»: los primeros 10 s tras arrancar. «Templado»: los 20 s siguientes.

| variante | peticiones/s en frío | peticiones/s templado | p50 | p99 | memoria máx. bajo carga |
|---|---|---|---|---|---|
| go | 466 | 468 | 26.9 ms | 123.0 ms | 32 MiB |
| go, GOGC=400 | 578 | 567 | 25.3 ms | 84.3 ms | 43 MiB |
| rust (698) | 836 | 831 | 19.0 ms | 30.8 ms | 11 MiB |
| rust tope (LTO) | 851 | 839 | 18.8 ms | 30.5 ms | 11 MiB |
| rust + mimalloc | 1425 | 1448 | 10.8 ms | 17.9 ms | 26 MiB |
| rust musl (698) | 187 | 178 | 89.1 ms | 156.8 ms | 9 MiB |
| rust musl + mimalloc | 693 | 691 | 22.8 ms | 36.8 ms | 28 MiB |
| rust musl + jemalloc | 643 | 644 | 24.4 ms | 39.6 ms | 11 MiB |
| actix-web | 834 | 822 | 19.2 ms | 23.3 ms | 11 MiB |
| rust debug | 192 | 196 | 80.7 ms | 128.6 ms | 10 MiB |

## 5 · Carga de CPU pura: /suma, a tope

`hey -c 16` desde otro contenedor; servidor con 2 CPU y 1g. «En frío»: los primeros 10 s tras arrancar. «Templado»: los 20 s siguientes.

| variante | peticiones/s en frío | peticiones/s templado | p50 | p99 | memoria máx. bajo carga |
|---|---|---|---|---|---|
| go | 14421 | 13598 | 1.0 ms | 3.3 ms | 13 MiB |
| go, GOGC=400 | 13055 | 14964 | 1.0 ms | 2.6 ms | 24 MiB |
| rust (698) | 16587 | 16130 | 1.0 ms | 1.9 ms | 5 MiB |
| rust tope (LTO) | 16281 | 16735 | 0.9 ms | 1.6 ms | 5 MiB |
| rust + mimalloc | 16668 | 16638 | 0.9 ms | 1.6 ms | 11 MiB |
| rust musl (698) | 15882 | 15470 | 1.0 ms | 2.1 ms | 4 MiB |
| rust musl + mimalloc | 15648 | 15755 | 1.0 ms | 2.1 ms | 13 MiB |
| rust musl + jemalloc | 15782 | 16238 | 1.0 ms | 1.9 ms | 5 MiB |
| actix-web | 16726 | 16711 | 0.9 ms | 1.3 ms | 5 MiB |
