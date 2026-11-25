# SALIDA — ¿Go o Rust? (golandia 698), con el banco de la revancha

> Corrida del 2026-10-29 07:53. Regenerable con `python canales/golandia/recursos/pruebas-go-vs-rust/ejecutar_todo.py`. ⛔ Ninguna cifra de blog: todo salió de aquí.

- go: `go version go1.27.2 linux/amd64` (net/http, solo librería estándar)
- rust: `rustc 1.99.1 (58cf8f9c3 2026-10-13)` — axum 0.8.9 sobre tokio 1.54.1 (versiones fijadas con `=` en Cargo.toml y Cargo.lock)
- Docker `29.6.2`. Servidores con 2 CPU y 1g; `hey -c 16` con 4 CPU desde otro contenedor.
- otros contenedores encendidos al empezar (de otros agentes): 1 ['v699-srv']

## 1 · Construir la imagen desde cero (`docker build --no-cache`, dependencias incluidas)

| lenguaje | qué se cronometra | segundos |
|---|---|---|
| go | go build | 5.3 |
| rust (glibc) | cargo build --release (baja y compila axum, tokio y 50 paquetes más) | 10.4 |
| rust (musl) | lo mismo, en rust:1.99.1-alpine | 22.5 |

## 2 · Lo que despliegas

| imagen | base | tamaño |
|---|---|---|
| go | scratch + binario estático | 2.4 MiB |
| rust (glibc) | debian:trixie-slim + binario (Rust 1.99.1, axum 0.8.9, tokio 1.54.1) | 29.0 MiB |
| rust (musl) | scratch + binario estático musl (Rust 1.99.1, axum 0.8.9, tokio 1.54.1) | 0.6 MiB |

`docker image inspect -f {{.Size}}`, como en 651-654. Con el almacén de imágenes de containerd (Docker 29) es el tamaño comprimido, lo que se descarga: por eso la imagen de Go (2,4 MiB) es más chica que su binario (sección 9).

## 0 · ¿Hacen lo mismo? (antes de medir)

- go: total 9990000, primeros ids [0, 1000, 2000, 3000, 4000]…, suma 99863963591 → igual que Go
- rust (glibc): total 9990000, primeros ids [0, 1000, 2000, 3000, 4000]…, suma 99863963591 → igual que Go
- rust (musl): total 9990000, primeros ids [0, 1000, 2000, 3000, 4000]…, suma 99863963591 → igual que Go

## 3 · Arranque: contenedor nuevo → primera respuesta 200 (mediana de 8) y memoria en reposo

| variante | 1ª respuesta (docker incluido) | «listo» (su propio reloj) | memoria en reposo |
|---|---|---|---|
| go | 206 ms | 0.0 ms | 2.8 MiB |
| rust (glibc) | 205 ms | 0.0 ms | 2.0 MiB |
| rust (musl) | 216 ms | 0.1 ms | 1.8 MiB |

## 4 · Carga con memoria: /trabajo (objetos, mapa, orden, JSON), a tope

`hey -c 16` desde otro contenedor; servidor con 2 CPU y 1g. «En frío»: los primeros 10 s tras arrancar. «Templado»: los 20 s siguientes.

| variante | peticiones/s en frío | peticiones/s templado | p50 | p99 | memoria máx. bajo carga |
|---|---|---|---|---|---|
| go | 438 | 427 | 30.9 ms | 121.8 ms | 30 MiB |
| rust (glibc) | 784 | 754 | 20.9 ms | 36.2 ms | 11 MiB |
| rust (musl) | 178 | 174 | 91.3 ms | 160.0 ms | 10 MiB |

## 5 · Carga de CPU pura: /suma, a tope

`hey -c 16` desde otro contenedor; servidor con 2 CPU y 1g. «En frío»: los primeros 10 s tras arrancar. «Templado»: los 20 s siguientes.

| variante | peticiones/s en frío | peticiones/s templado | p50 | p99 | memoria máx. bajo carga |
|---|---|---|---|---|---|
| go | 13696 | 13501 | 1.1 ms | 3.3 ms | 10 MiB |
| rust (glibc) | 15424 | 15468 | 1.0 ms | 2.1 ms | 2 MiB |
| rust (musl) | 14923 | 15510 | 1.0 ms | 2.1 ms | 5 MiB |

## 6 · Latencia de una petición sola: /trabajo

Un solo cliente a 5 peticiones/s (`hey -c 1 -q 5`), 20 s, tras 10 s de calentamiento a tope. Cada petición llega sola: no hay cola, es lo que tarda el servidor en atender una.

| variante | peticiones/s logradas | p50 | p99 |
|---|---|---|---|
| go | 5 | 4.00 ms | 8.80 ms |
| rust (glibc) | 5 | 3.40 ms | 4.00 ms |
| rust (musl) | 5 | 5.50 ms | 6.70 ms |

## 7 · Latencia de una petición sola: /suma

Un solo cliente a 5 peticiones/s (`hey -c 1 -q 5`), 20 s, tras 10 s de calentamiento a tope. Cada petición llega sola: no hay cola, es lo que tarda el servidor en atender una.

| variante | peticiones/s logradas | p50 | p99 |
|---|---|---|---|
| go | 5 | 0.50 ms | 1.00 ms |
| rust (glibc) | 5 | 0.50 ms | 1.20 ms |
| rust (musl) | 5 | 0.50 ms | 1.40 ms |

## 8 · Cómputo puro (reloj del propio proceso, sin red)

Como `investigacion/pruebas_go_vs_x_computo.py`: contenedor con 2 CPU y 1g, 50 peticiones de calentamiento y 300 medidas, de una en una; el tiempo es el campo «ns» que mide cada servidor alrededor de su cálculo.

| variante | /suma?n=200000 mediana | p99 | /trabajo?n=20000 mediana | p99 |
|---|---|---|---|---|
| go | 0.122 ms | 0.538 ms | 3.422 ms | 3.971 ms |
| rust (glibc) | 0.111 ms | 0.287 ms | 1.884 ms | 2.346 ms |
| rust (musl) | 0.111 ms | 0.137 ms | 4.043 ms | 4.896 ms |

## 9 · Lo que cuesta escribirlo

| | go | rust |
|---|---|---|
| líneas de código (sin blancos ni comentarios) | 90 | 67 |
| líneas totales del archivo | 102 | 79 |
| dependencias externas | 0 (solo librería estándar) | 2 directas (axum, tokio) · 51 paquetes en Cargo.lock |

Compilación dentro de `golang:1.27.2` y `rust:1.99.1-trixie` con `--cpus 2`, contenedor nuevo cada vez (cachés vacías), mediana de 3. «Incremental»: cambiar una cadena en el manejador de /espera y volver a compilar.

| paso | go (s) | rust (s) | las 3 corridas go · rust |
|---|---|---|---|
| bajar dependencias | — | 1.5 | — · 3.1 / 1.5 / 1.1 |
| compilación limpia (release) | 7.9 | 12.9 | 7.9 / 8.0 / 7.9 · 12.9 / 13.1 / 12.4 |
| incremental (release) | 0.17 | 1.42 | 0.17 / 0.16 / 0.17 · 1.42 / 1.46 / 1.39 |
| compilación limpia (debug) | (Go no tiene modo aparte) | 8.5 | — · 8.7 / 8.4 / 8.5 |
| incremental (debug) | — | 0.33 | — · 0.33 / 0.38 / 0.30 |

| binario | bytes | MiB |
|---|---|---|
| go (`-ldflags='-s -w'`) | 5873824 | 5.60 |
| rust glibc (`strip = true`) | 1282400 | 1.22 |
| rust musl (estático, el de la imagen scratch) | 1397832 | 1.33 |

## Anexo · los núcleos que ve tokio dentro del contenedor

- `docker run --cpus 2 v698-govs-rust:1` → `LISTO_EN_MS 0.046 rust hilos=2`
- `docker run v698-govs-rust:1` → `LISTO_EN_MS 0.057 rust hilos=24`

`hilos` es `std::thread::available_parallelism()`, el número con el que `#[tokio::main]` crea sus hilos de trabajo.

- otros contenedores encendidos al terminar: 0
