# SALIDA — ¿Go o Rust? La revancha (golandia 958), corrida 3, con el banco de la serie

> Corrida del 2026-11-25 08:43, con el candado `_logs/banco.lock` tomado. Regenerable con `python investigacion/pruebas_go_rust_revancha.py 3` (mismo banco que 879, 865 y 902-905; el 698 es `recursos/pruebas-go-vs-rust/SALIDA.md`). ⛔ Ninguna cifra de blog: todo salió de aquí.

- go: `go version go1.27.2 linux/amd64` (net/http, solo librería estándar)
- rust: `rustc 1.99.1 (58cf8f9c3 2026-10-13) · cargo 1.99.1 (5f94df478 2026-08-27)` · alpine `host: x86_64-unknown-linux-musl`
- crates (Cargo.lock): axum 0.8.9 · tokio 1.54.1 · mimalloc 0.1.52 · tikv-jemallocator 0.6.1 · actix-web 4.20.0 (su tokio 1.54.1)
- Docker `29.6.2`. Servidores con 2 CPU y 1g; `hey -c 16` con 4 CPU desde otro contenedor.

## 1 · Construir la imagen desde cero (`docker build --no-cache`, dependencias incluidas)

| variante | qué se cronometra | segundos |
|---|---|---|
| go | go build | 5.5 |
| rust (698) | cargo build --release (axum, glibc) | 12.6 |
| rust tope | cargo build --profile tope (LTO fat, codegen-units=1, panic=abort) | 15.0 |
| rust + mimalloc | cargo build --release --features mimalloc (glibc) | 10.9 |
| rust musl (698) | cargo build --release en rust:1.99.1-alpine | 21.6 |
| rust musl + mimalloc | lo mismo con --features mimalloc | 20.3 |
| rust musl + jemalloc | apk add make + --features jemalloc (compila jemalloc) | 27.9 |
| actix-web | cargo build --release (actix-web 4, glibc) | 21.7 |
| rust debug | cargo build (sin --release) | 16.6 |

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

- CPU de Windows (Win32_Processor.LoadPercentage) antes de medir: 15 %
## Estado de la máquina antes de medir

- /proc/loadavg (desde un contenedor alpine, VM de Docker con 24 hilos): `2.36 2.54 1.11 4/541 6`
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
| go | 209 ms | 0.0 ms | 2.6 MiB |
| go, GOGC=400 | 202 ms | 0.0 ms | 2.7 MiB |
| rust (698) | 216 ms | 0.1 ms | 1.8 MiB |
| rust tope (LTO) | 204 ms | 0.0 ms | 1.8 MiB |
| rust + mimalloc | 207 ms | 0.1 ms | 7.9 MiB |
| rust musl (698) | 206 ms | 0.1 ms | 1.9 MiB |
| rust musl + mimalloc | 213 ms | 0.1 ms | 8.0 MiB |
| rust musl + jemalloc | 205 ms | 0.1 ms | 2.3 MiB |
| actix-web | 207 ms | 0.1 ms | 2.3 MiB |
| rust debug | 205 ms | 0.1 ms | 2.3 MiB |

## 3b · Lo que dice cada servidor al arrancar (con --cpus 2)

- go → `LISTO_EN_MS 0.047`
- go, GOGC=400 → `LISTO_EN_MS 0.039`
- rust (698) → `LISTO_EN_MS 0.055 rust axum hilos=2 repartidor=sistema perfil=optimizado`
- rust tope (LTO) → `LISTO_EN_MS 0.057 rust axum hilos=2 repartidor=sistema perfil=optimizado`
- rust + mimalloc → `LISTO_EN_MS 0.057 rust axum hilos=2 repartidor=mimalloc perfil=optimizado`
- rust musl (698) → `LISTO_EN_MS 0.072 rust axum hilos=2 repartidor=sistema perfil=optimizado`
- rust musl + mimalloc → `LISTO_EN_MS 0.070 rust axum hilos=2 repartidor=mimalloc perfil=optimizado`
- rust musl + jemalloc → `LISTO_EN_MS 0.056 rust axum hilos=2 repartidor=jemalloc perfil=optimizado`
- actix-web → `LISTO_EN_MS 0.069 rust actix hilos=2`
- rust debug → `LISTO_EN_MS 0.139 rust axum hilos=2 repartidor=sistema perfil=debug`

## 4 · Carga con memoria: /trabajo (objetos, mapa, orden, JSON), a tope

`hey -c 16` desde otro contenedor; servidor con 2 CPU y 1g. «En frío»: los primeros 10 s tras arrancar. «Templado»: los 20 s siguientes.

| variante | peticiones/s en frío | peticiones/s templado | p50 | p99 | memoria máx. bajo carga |
|---|---|---|---|---|---|
| go | 436 | 435 | 29.8 ms | 127.5 ms | 33 MiB |
| go, GOGC=400 | 542 | 536 | 27.0 ms | 87.0 ms | 44 MiB |
| rust (698) | 759 | 764 | 20.9 ms | 33.4 ms | 11 MiB |
| rust tope (LTO) | 790 | 789 | 20.2 ms | 32.6 ms | 11 MiB |
| rust + mimalloc | 1355 | 1346 | 11.9 ms | 19.8 ms | 30 MiB |
| rust musl (698) | 159 | 167 | 95.4 ms | 163.2 ms | 9 MiB |
| rust musl + mimalloc | 664 | 669 | 23.6 ms | 37.9 ms | 27 MiB |
| rust musl + jemalloc | 624 | 616 | 25.6 ms | 41.8 ms | 11 MiB |
| actix-web | 785 | 775 | 20.4 ms | 24.5 ms | 11 MiB |
| rust debug | 179 | 180 | 87.8 ms | 138.3 ms | 10 MiB |

## 5 · Carga de CPU pura: /suma, a tope

`hey -c 16` desde otro contenedor; servidor con 2 CPU y 1g. «En frío»: los primeros 10 s tras arrancar. «Templado»: los 20 s siguientes.

| variante | peticiones/s en frío | peticiones/s templado | p50 | p99 | memoria máx. bajo carga |
|---|---|---|---|---|---|
| go | 13996 | 13981 | 1.0 ms | 3.1 ms | 12 MiB |
| go, GOGC=400 | 13951 | 13937 | 1.0 ms | 3.1 ms | 22 MiB |
| rust (698) | 15487 | 15582 | 1.0 ms | 2.1 ms | 3 MiB |
| rust tope (LTO) | 15862 | 15873 | 1.0 ms | 2.0 ms | 4 MiB |
| rust + mimalloc | 15584 | 15540 | 1.0 ms | 2.2 ms | 12 MiB |
| rust musl (698) | 15130 | 15212 | 1.0 ms | 2.2 ms | 4 MiB |
| rust musl + mimalloc | 15546 | 15569 | 1.0 ms | 2.1 ms | 13 MiB |
| rust musl + jemalloc | 15288 | 15499 | 1.0 ms | 2.2 ms | 6 MiB |
| actix-web | 15770 | 15763 | 1.0 ms | 2.0 ms | 5 MiB |
| rust debug | 2695 | 2690 | 5.8 ms | 10.1 ms | 4 MiB |

## 6 · Latencia de una petición sola: /trabajo

Un solo cliente a 5 peticiones/s (`hey -c 1 -q 5`), 20 s, tras 10 s de calentamiento a tope. Cada petición llega sola: no hay cola, es lo que tarda el servidor en atender una.

| variante | peticiones/s logradas | p50 | p99 |
|---|---|---|---|
| go | 5 | 4.00 ms | 5.30 ms |
| go, GOGC=400 | 5 | 3.50 ms | 5.60 ms |
| rust (698) | 5 | 3.70 ms | 5.00 ms |
| rust tope (LTO) | 5 | 3.10 ms | 5.00 ms |
| rust + mimalloc | 5 | 2.30 ms | 3.00 ms |
| rust musl (698) | 5 | 5.30 ms | 6.40 ms |
| rust musl + mimalloc | 5 | 3.40 ms | 4.50 ms |
| rust musl + jemalloc | 5 | 3.60 ms | 4.50 ms |
| actix-web | 5 | 3.20 ms | 5.00 ms |
| rust debug | 5 | 11.70 ms | 14.60 ms |

## 7 · Latencia de una petición sola: /suma

Un solo cliente a 5 peticiones/s (`hey -c 1 -q 5`), 20 s, tras 10 s de calentamiento a tope. Cada petición llega sola: no hay cola, es lo que tarda el servidor en atender una.

| variante | peticiones/s logradas | p50 | p99 |
|---|---|---|---|
| go | 5 | 0.60 ms | 2.10 ms |
| go, GOGC=400 | 5 | 0.50 ms | 1.80 ms |
| rust (698) | 5 | 0.50 ms | 1.20 ms |
| rust tope (LTO) | 5 | 0.50 ms | 1.10 ms |
| rust + mimalloc | 5 | 0.50 ms | 1.30 ms |
| rust musl (698) | 5 | 0.50 ms | 1.30 ms |
| rust musl + mimalloc | 5 | 0.50 ms | 1.30 ms |
| rust musl + jemalloc | 5 | 0.50 ms | 1.40 ms |
| actix-web | 5 | 0.50 ms | 1.20 ms |
| rust debug | 5 | 1.20 ms | 2.00 ms |

## 8 · Cómputo puro: el reloj del propio proceso, sin red

50 peticiones de calentamiento y 300 medidas, de una en una; el tiempo es el campo «ns» que mide cada servidor alrededor de su propio cálculo.

| variante | /suma?n=200000 mediana | p99 | /trabajo?n=20000 mediana | p99 |
|---|---|---|---|---|
| go | 0.122 ms | 0.272 ms | 3.422 ms | 4.692 ms |
| go, GOGC=400 | 0.122 ms | 0.180 ms | 2.939 ms | 4.021 ms |
| rust (698) | 0.111 ms | 0.229 ms | 1.909 ms | 2.427 ms |
| rust tope (LTO) | 0.109 ms | 0.230 ms | 1.839 ms | 2.811 ms |
| rust + mimalloc | 0.111 ms | 0.229 ms | 1.508 ms | 2.234 ms |
| rust musl (698) | 0.111 ms | 0.235 ms | 4.261 ms | 6.011 ms |
| rust musl + mimalloc | 0.111 ms | 0.203 ms | 2.805 ms | 3.342 ms |
| rust musl + jemalloc | 0.109 ms | 0.229 ms | 2.977 ms | 3.765 ms |
| actix-web | 0.111 ms | 0.229 ms | 2.095 ms | 2.943 ms |
| rust debug | 0.660 ms | 0.954 ms | 9.421 ms | 12.404 ms |

## 9 · Muchas conexiones esperando a la vez: /espera?ms=1000

Cada petición duerme 1000 ms sin gastar CPU (`time.Sleep` en Go; `tokio::time::sleep` en axum y en actix). Con N clientes el techo teórico es N × 1 peticiones/s: aquí no se mide cálculo, se mide cuánto cuesta tener N conexiones abiertas a la vez. `hey -c N -z 20s -t 30` desde otro contenedor con 4 CPU, tras 10 s de calentamiento con los mismos clientes; servidor con 2 CPU y 1 GiB. Memoria máxima con `docker stats` durante la medida.

| variante | clientes | peticiones/s | p50 | p99 | errores | memoria máx. |
|---|---|---|---|---|---|---|
| go | 1000 | 992 | 1001.3 ms | 1119.3 ms | 0 (de 20000) | 50 MiB |
| go | 4000 | 3940 | 1001.4 ms | 1196.9 ms | 0 (de 80000) | 190 MiB |
| rust (698) | 1000 | 992 | 1001.8 ms | 1100.9 ms | 0 (de 20000) | 35 MiB |
| rust (698) | 4000 | 3885 | 1002.1 ms | 1176.7 ms | 0 (de 79343) | 130 MiB |
| actix-web | 1000 | 993 | 1002.1 ms | 1083.1 ms | 0 (de 20000) | 27 MiB |
| actix-web | 4000 | 3939 | 1002.7 ms | 1178.9 ms | 0 (de 80000) | 108 MiB |

## 10 · Lo que cuesta compilarlo

| | go | rust axum | rust actix-web |
|---|---|---|---|
| líneas de código (sin blancos ni comentarios) | 90 | 75 (el 698: 67; lo demás es el repartidor) | 68 |
| paquetes en Cargo.lock | 0 (solo librería estándar) | 58 (incluye mimalloc y jemalloc opcionales) | 112 |

Compilación dentro de `golang:1.27.2` y `rust:1.99.1-trixie` con `--cpus 2`, contenedor nuevo cada vez (cachés vacías), mediana de 3. «Incremental»: cambiar una cadena en el manejador de /espera y volver a compilar. «tope» = LTO fat + codegen-units=1 + panic=abort.

| paso | segundos (mediana) | las 3 corridas |
|---|---|---|
| go: compilación limpia | 7.2 | 7.4 / 7.2 / 7.1 |
| go: incremental | 0.14 | 0.16 / 0.13 / 0.14 |
| rust: bajar dependencias | 1.8 | 2.8 / 1.8 / 1.3 |
| rust release: compilación limpia | 11.7 | 11.7 / 12.2 / 11.5 |
| rust release: incremental | 1.31 | 1.31 / 1.33 / 1.31 |
| rust tope (LTO): compilación limpia | 12.7 | 12.8 / 12.6 / 12.7 |
| rust tope (LTO): incremental | 4.82 | 4.82 / 4.82 / 4.76 |
| rust debug: compilación limpia | 7.5 | 7.5 / 7.5 / 7.6 |
| rust debug: incremental | 0.30 | 0.30 / 0.37 / 0.30 |

Al terminar: /proc/loadavg `1.43 0.99 1.20 18/517 6`; CPU de Windows 2 %; ajenos: ninguno

