# SALIDA — ¿Go o Rust? La revancha (golandia 958), corrida humo, con el banco de la serie

> Corrida del 2026-11-24 21:05, con el candado `_logs/banco.lock` tomado. Regenerable con `python investigacion/pruebas_go_rust_revancha.py humo` (mismo banco que 879, 865 y 902-905; el 698 es `recursos/pruebas-go-vs-rust/SALIDA.md`). ⛔ Ninguna cifra de blog: todo salió de aquí.

- go: `go version go1.27.2 linux/amd64` (net/http, solo librería estándar)
- rust: `rustc 1.99.1 (58cf8f9c3 2026-10-13) · cargo 1.99.1 (5f94df478 2026-08-27)` · alpine `host: x86_64-unknown-linux-musl`
- crates (Cargo.lock): axum 0.8.9 · tokio 1.54.1 · mimalloc 0.1.52 · tikv-jemallocator 0.6.1 · actix-web 4.20.0 (su tokio 1.54.1)
- Docker `29.6.2`. Servidores con 2 CPU y 1g; `hey -c 16` con 4 CPU desde otro contenedor.

## 1 · Construir la imagen desde cero (`docker build --no-cache`, dependencias incluidas)

| variante | qué se cronometra | segundos |
|---|---|---|

## 2 · Lo que despliegas

| imagen | base | tamaño |
|---|---|---|
| go | scratch + binario estático | 2.4 MiB |
| rust | debian:trixie-slim + binario release (el del 698) | 29.0 MiB |

| binario (lo que se copia a la imagen) | bytes | MiB |
|---|---|---|
| go (`-ldflags='-s -w'`) | 5873824 | 5.60 |
| rust | 1282944 | 1.22 |

- CPU de Windows (Win32_Processor.LoadPercentage) antes de medir: 25 %
## 0 · ¿Hacen lo mismo? (antes de medir)

- go: total 9990000, primeros ids [0, 1000, 2000, 3000, 4000]…, suma 99863963591 → igual que Go
- rust (698): total 9990000, primeros ids [0, 1000, 2000, 3000, 4000]…, suma 99863963591 → igual que Go

## 3 · Arranque: contenedor nuevo → primera respuesta 200 (mediana de 1) y memoria en reposo

| variante | 1ª respuesta (docker incluido) | «listo» (su propio reloj) | memoria en reposo |
|---|---|---|---|
| go | 198 ms | 0.1 ms | 2.7 MiB |
| rust (698) | 209 ms | 0.1 ms | 2.0 MiB |

## 3b · Lo que dice cada servidor al arrancar (con --cpus 2)

- go → `LISTO_EN_MS 0.080`
- rust (698) → `LISTO_EN_MS 0.054 rust axum hilos=2 repartidor=sistema perfil=optimizado`

## 4 · Carga con memoria: /trabajo (objetos, mapa, orden, JSON), a tope

`hey -c 16` desde otro contenedor; servidor con 2 CPU y 1g. «En frío»: los primeros 2 s tras arrancar. «Templado»: los 3 s siguientes.

| variante | peticiones/s en frío | peticiones/s templado | p50 | p99 | memoria máx. bajo carga |
|---|---|---|---|---|---|
| go | 430 | 428 | 30.1 ms | 128.0 ms | 28 MiB |
| rust (698) | 763 | 767 | 20.8 ms | 33.4 ms | 10 MiB |

## 5 · Carga de CPU pura: /suma, a tope

`hey -c 16` desde otro contenedor; servidor con 2 CPU y 1g. «En frío»: los primeros 2 s tras arrancar. «Templado»: los 3 s siguientes.

| variante | peticiones/s en frío | peticiones/s templado | p50 | p99 | memoria máx. bajo carga |
|---|---|---|---|---|---|
| go | 13350 | 13328 | 1.1 ms | 3.5 ms | 10 MiB |
| rust (698) | 15523 | 15689 | 1.0 ms | 2.0 ms | 4 MiB |

## 9 · Muchas conexiones esperando a la vez: /espera?ms=1000

Cada petición duerme 1000 ms sin gastar CPU (`time.Sleep` en Go; `tokio::time::sleep` en axum y en actix). Con N clientes el techo teórico es N × 1 peticiones/s: aquí no se mide cálculo, se mide cuánto cuesta tener N conexiones abiertas a la vez. `hey -c N -z 3s -t 30` desde otro contenedor con 4 CPU, tras 2 s de calentamiento con los mismos clientes; servidor con 2 CPU y 1 GiB. Memoria máxima con `docker stats` durante la medida.

| variante | clientes | peticiones/s | p50 | p99 | errores | memoria máx. |
|---|---|---|---|---|---|---|
| go | 1000 | 959 | 1004.7 ms | 1115.1 ms | 0 (de 3000) | 46 MiB |
| go | 4000 | 3682 | 1009.3 ms | 1221.9 ms | 0 (de 12000) | 162 MiB |
| rust (698) | 1000 | 956 | 1002.3 ms | 1125.1 ms | 0 (de 3000) | 32 MiB |
| rust (698) | 4000 | 3297 | 1006.0 ms | 2145.4 ms | 0 (de 11162) | 121 MiB |

## 10 · Lo que cuesta compilarlo

| | go | rust axum | rust actix-web |
|---|---|---|---|
| líneas de código (sin blancos ni comentarios) | 90 | 75 (con las 10 del repartidor) | 68 |
| paquetes en Cargo.lock | 0 (solo librería estándar) | 58 (incluye mimalloc y jemalloc opcionales) | 112 |

Compilación dentro de `golang:1.27.2` y `rust:1.99.1-trixie` con `--cpus 2`, contenedor nuevo cada vez (cachés vacías), mediana de 1. «Incremental»: cambiar una cadena en el manejador de /espera y volver a compilar. «tope» = LTO fat + codegen-units=1 + panic=abort.

| paso | segundos (mediana) | las 3 corridas |
|---|---|---|
| go: compilación limpia | 9.1 | 9.1 |
| go: incremental | 0.18 | 0.18 |
| rust: bajar dependencias | 2.7 | 2.7 |
| rust release: compilación limpia | 13.3 | 13.3 |
| rust release: incremental | 1.46 | 1.46 |
| rust tope (LTO): compilación limpia | 14.7 | 14.7 |
| rust tope (LTO): incremental | 5.49 | 5.49 |
| rust debug: compilación limpia | 8.7 | 8.7 |
| rust debug: incremental | 0.30 | 0.30 |

Al terminar: /proc/loadavg `1.70 2.38 1.28 7/502 6`; CPU de Windows 13 %; ajenos: ninguno

