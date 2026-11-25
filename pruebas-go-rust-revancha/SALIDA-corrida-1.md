# SALIDA — ¿Go o Rust? La revancha (golandia 958), corrida 1, con el banco de la serie

> Corrida del 2026-11-24 21:08, con el candado `_logs/banco.lock` tomado. Regenerable con `python investigacion/pruebas_go_rust_revancha.py 1` (mismo banco que 879, 865 y 902-905; el 698 es `recursos/pruebas-go-vs-rust/SALIDA.md`). ⛔ Ninguna cifra de blog: todo salió de aquí.

- go: `go version go1.27.2 linux/amd64` (net/http, solo librería estándar)
- rust: `rustc 1.99.1 (58cf8f9c3 2026-10-13) · cargo 1.99.1 (5f94df478 2026-08-27)` · alpine `host: x86_64-unknown-linux-musl`
- crates (Cargo.lock): axum 0.8.9 · tokio 1.54.1 · mimalloc 0.1.52 · tikv-jemallocator 0.6.1 · actix-web 4.20.0 (su tokio 1.54.1)
- Docker `29.6.2`. Servidores con 2 CPU y 1g; `hey -c 16` con 4 CPU desde otro contenedor.

## 1 · Construir la imagen desde cero (`docker build --no-cache`, dependencias incluidas)

| variante | qué se cronometra | segundos |
|---|---|---|
| go | go build | 6.2 |
| rust (698) | cargo build --release (axum, glibc) | 12.0 |
| rust tope | cargo build --profile tope (LTO fat, codegen-units=1, panic=abort) | 17.3 |
| rust + mimalloc | cargo build --release --features mimalloc (glibc) | 12.9 |
| rust musl (698) | cargo build --release en rust:1.99.1-alpine | 22.7 |
| rust musl + mimalloc | lo mismo con --features mimalloc | 24.1 |
| rust musl + jemalloc | apk add make + --features jemalloc (compila jemalloc) | 30.7 |
| actix-web | cargo build --release (actix-web 4, glibc) | 21.1 |
| rust debug | cargo build (sin --release) | 15.7 |

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

- CPU de Windows (Win32_Processor.LoadPercentage) antes de medir: 23 %
## Estado de la máquina antes de medir

- /proc/loadavg (desde un contenedor alpine, VM de Docker con 24 hilos): `2.99 3.04 1.78 3/500 6`
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
| go | 213 ms | 0.0 ms | 2.6 MiB |
| go, GOGC=400 | 197 ms | 0.0 ms | 2.9 MiB |
| rust (698) | 201 ms | 0.1 ms | 2.0 MiB |
| rust tope (LTO) | 209 ms | 0.0 ms | 2.0 MiB |
| rust + mimalloc | 208 ms | 0.1 ms | 8.1 MiB |
| rust musl (698) | 202 ms | 0.1 ms | 1.9 MiB |
| rust musl + mimalloc | 204 ms | 0.1 ms | 7.9 MiB |
| rust musl + jemalloc | 215 ms | 0.1 ms | 2.3 MiB |
| actix-web | 199 ms | 0.1 ms | 2.2 MiB |
| rust debug | 207 ms | 0.1 ms | 2.3 MiB |

## 3b · Lo que dice cada servidor al arrancar (con --cpus 2)

- go → `LISTO_EN_MS 0.056`
- go, GOGC=400 → `LISTO_EN_MS 0.045`
- rust (698) → `LISTO_EN_MS 0.068 rust axum hilos=2 repartidor=sistema perfil=optimizado`
- rust tope (LTO) → `LISTO_EN_MS 0.092 rust axum hilos=2 repartidor=sistema perfil=optimizado`
- rust + mimalloc → `LISTO_EN_MS 0.304 rust axum hilos=2 repartidor=mimalloc perfil=optimizado`
- rust musl (698) → `LISTO_EN_MS 0.196 rust axum hilos=2 repartidor=sistema perfil=optimizado`
- rust musl + mimalloc → `LISTO_EN_MS 0.060 rust axum hilos=2 repartidor=mimalloc perfil=optimizado`
- rust musl + jemalloc → `LISTO_EN_MS 0.086 rust axum hilos=2 repartidor=jemalloc perfil=optimizado`
- actix-web → `LISTO_EN_MS 0.067 rust actix hilos=2`
- rust debug → `LISTO_EN_MS 0.199 rust axum hilos=2 repartidor=sistema perfil=debug`

## 4 · Carga con memoria: /trabajo (objetos, mapa, orden, JSON), a tope

`hey -c 16` desde otro contenedor; servidor con 2 CPU y 1g. «En frío»: los primeros 10 s tras arrancar. «Templado»: los 20 s siguientes.

| variante | peticiones/s en frío | peticiones/s templado | p50 | p99 | memoria máx. bajo carga |
|---|---|---|---|---|---|
| go | 425 | 426 | 30.7 ms | 128.1 ms | 31 MiB |
| go, GOGC=400 | 531 | 533 | 27.1 ms | 86.4 ms | 50 MiB |
| rust (698) | 763 | 762 | 20.9 ms | 33.7 ms | 12 MiB |
| rust tope (LTO) | 789 | 771 | 20.6 ms | 33.4 ms | 11 MiB |
| rust + mimalloc | 1349 | 1355 | 11.8 ms | 19.7 ms | 32 MiB |
| rust musl (698) | 169 | 162 | 99.0 ms | 165.8 ms | 9 MiB |
| rust musl + mimalloc | 668 | 663 | 23.8 ms | 38.6 ms | 28 MiB |
| rust musl + jemalloc | 618 | 618 | 25.5 ms | 41.4 ms | 11 MiB |
| actix-web | 788 | 786 | 20.1 ms | 23.9 ms | 11 MiB |
| rust debug | 180 | 180 | 87.7 ms | 143.0 ms | 10 MiB |

## 5 · Carga de CPU pura: /suma, a tope

`hey -c 16` desde otro contenedor; servidor con 2 CPU y 1g. «En frío»: los primeros 10 s tras arrancar. «Templado»: los 20 s siguientes.

| variante | peticiones/s en frío | peticiones/s templado | p50 | p99 | memoria máx. bajo carga |
|---|---|---|---|---|---|
| go | 13832 | 13845 | 1.0 ms | 3.3 ms | 10 MiB |
| go, GOGC=400 | 13935 | 13858 | 1.0 ms | 3.3 ms | 23 MiB |
| rust (698) | 15539 | 15646 | 1.0 ms | 2.1 ms | 4 MiB |
| rust tope (LTO) | 15864 | 15760 | 1.0 ms | 2.1 ms | 5 MiB |
| rust + mimalloc | 15456 | 15557 | 1.0 ms | 2.2 ms | 12 MiB |
| rust musl (698) | 15098 | 15168 | 1.0 ms | 2.4 ms | 3 MiB |
| rust musl + mimalloc | 15665 | 15114 | 1.0 ms | 2.8 ms | 13 MiB |
| rust musl + jemalloc | 15551 | 15668 | 1.0 ms | 2.2 ms | 5 MiB |
| actix-web | 15095 | 15782 | 1.0 ms | 2.0 ms | 5 MiB |
| rust debug | 2725 | 2705 | 5.8 ms | 10.1 ms | 3 MiB |

## 6 · Latencia de una petición sola: /trabajo

Un solo cliente a 5 peticiones/s (`hey -c 1 -q 5`), 20 s, tras 10 s de calentamiento a tope. Cada petición llega sola: no hay cola, es lo que tarda el servidor en atender una.

| variante | peticiones/s logradas | p50 | p99 |
|---|---|---|---|
| go | 5 | 4.10 ms | 6.60 ms |
| go, GOGC=400 | 5 | 3.40 ms | 5.90 ms |
| rust (698) | 5 | 3.10 ms | 3.90 ms |
| rust tope (LTO) | 5 | 3.00 ms | 4.50 ms |
| rust + mimalloc | 5 | 2.40 ms | 3.10 ms |
| rust musl (698) | 5 | 5.30 ms | 6.40 ms |
| rust musl + mimalloc | 5 | 3.30 ms | 4.30 ms |
| rust musl + jemalloc | 5 | 3.70 ms | 4.50 ms |
| actix-web | 5 | 3.40 ms | 4.70 ms |
| rust debug | 5 | 11.20 ms | 14.80 ms |

## 7 · Latencia de una petición sola: /suma

Un solo cliente a 5 peticiones/s (`hey -c 1 -q 5`), 20 s, tras 10 s de calentamiento a tope. Cada petición llega sola: no hay cola, es lo que tarda el servidor en atender una.

| variante | peticiones/s logradas | p50 | p99 |
|---|---|---|---|
| go | 5 | 0.50 ms | 1.60 ms |
| go, GOGC=400 | 5 | 0.50 ms | 1.20 ms |
| rust (698) | 5 | 0.50 ms | 1.20 ms |
| rust tope (LTO) | 5 | 0.50 ms | 1.30 ms |
| rust + mimalloc | 5 | 0.50 ms | 1.60 ms |
| rust musl (698) | 5 | 0.50 ms | 1.40 ms |
| rust musl + mimalloc | 5 | 0.60 ms | 1.60 ms |
| rust musl + jemalloc | 5 | 0.50 ms | 1.30 ms |
| actix-web | 5 | 0.60 ms | 1.40 ms |
| rust debug | 5 | 1.20 ms | 2.20 ms |

## 8 · Cómputo puro: el reloj del propio proceso, sin red

50 peticiones de calentamiento y 300 medidas, de una en una; el tiempo es el campo «ns» que mide cada servidor alrededor de su propio cálculo.

| variante | /suma?n=200000 mediana | p99 | /trabajo?n=20000 mediana | p99 |
|---|---|---|---|---|
| go | 0.122 ms | 0.202 ms | 3.377 ms | 4.305 ms |
| go, GOGC=400 | 0.122 ms | 0.222 ms | 2.933 ms | 3.860 ms |
| rust (698) | 0.111 ms | 0.158 ms | 1.863 ms | 2.837 ms |
| rust tope (LTO) | 0.109 ms | 0.165 ms | 1.792 ms | 2.853 ms |
| rust + mimalloc | 0.111 ms | 0.225 ms | 1.458 ms | 2.217 ms |
| rust musl (698) | 0.111 ms | 0.177 ms | 4.245 ms | 5.533 ms |
| rust musl + mimalloc | 0.111 ms | 0.195 ms | 2.834 ms | 3.331 ms |
| rust musl + jemalloc | 0.109 ms | 0.229 ms | 2.982 ms | 3.647 ms |
| actix-web | 0.111 ms | 0.182 ms | 2.069 ms | 2.510 ms |
| rust debug | 0.656 ms | 0.821 ms | 9.507 ms | 11.505 ms |

## 9 · Muchas conexiones esperando a la vez: /espera?ms=1000

Cada petición duerme 1000 ms sin gastar CPU (`time.Sleep` en Go; `tokio::time::sleep` en axum y en actix). Con N clientes el techo teórico es N × 1 peticiones/s: aquí no se mide cálculo, se mide cuánto cuesta tener N conexiones abiertas a la vez. `hey -c N -z 20s -t 30` desde otro contenedor con 4 CPU, tras 10 s de calentamiento con los mismos clientes; servidor con 2 CPU y 1 GiB. Memoria máxima con `docker stats` durante la medida.

| variante | clientes | peticiones/s | p50 | p99 | errores | memoria máx. |
|---|---|---|---|---|---|---|
| go | 1000 | 993 | 1001.5 ms | 1101.0 ms | 0 (de 20000) | 50 MiB |
| go | 4000 | 3936 | 1001.2 ms | 1231.0 ms | 0 (de 80000) | 182 MiB |
| rust (698) | 1000 | 990 | 1002.0 ms | 1122.7 ms | 0 (de 19990) | 36 MiB |
| rust (698) | 4000 | 3888 | 1002.2 ms | 2164.7 ms | 0 (de 79023) | 129 MiB |
| actix-web | 1000 | 991 | 1002.2 ms | 1133.2 ms | 0 (de 20000) | 28 MiB |
| actix-web | 4000 | 3946 | 1002.5 ms | 1192.7 ms | 0 (de 80000) | 109 MiB |

## 10 · Lo que cuesta compilarlo

| | go | rust axum | rust actix-web |
|---|---|---|---|
| líneas de código (sin blancos ni comentarios) | 90 | 75 (el 698: 67; lo demás es el repartidor) | 68 |
| paquetes en Cargo.lock | 0 (solo librería estándar) | 58 (incluye mimalloc y jemalloc opcionales) | 112 |

Compilación dentro de `golang:1.27.2` y `rust:1.99.1-trixie` con `--cpus 2`, contenedor nuevo cada vez (cachés vacías), mediana de 3. «Incremental»: cambiar una cadena en el manejador de /espera y volver a compilar. «tope» = LTO fat + codegen-units=1 + panic=abort.

| paso | segundos (mediana) | las 3 corridas |
|---|---|---|
| go: compilación limpia | 8.1 | 8.1 / 8.1 / 8.2 |
| go: incremental | 0.17 | 0.15 / 0.17 / 0.17 |
| rust: bajar dependencias | 2.9 | 3.2 / 2.9 / 2.0 |
| rust release: compilación limpia | 13.7 | 13.8 / 13.7 / 12.9 |
| rust release: incremental | 1.46 | 1.46 / 1.47 / 1.44 |
| rust tope (LTO): compilación limpia | 14.5 | 14.5 / 14.8 / 14.5 |
| rust tope (LTO): incremental | 5.48 | 5.48 / 5.55 / 5.47 |
| rust debug: compilación limpia | 8.7 | 8.7 / 8.9 / 8.7 |
| rust debug: incremental | 0.31 | 0.31 / 0.31 / 0.30 |

Al terminar: /proc/loadavg `1.83 1.54 1.59 10/502 6`; CPU de Windows 23 %; ajenos: ninguno

