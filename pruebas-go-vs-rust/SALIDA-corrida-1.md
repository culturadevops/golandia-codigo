> Nota al archivarla (2026-10-29): primera corrida, guardada para ver que se repite. La fila «rust musl» de la tabla de binarios mide la imagen comprimida, no el binario (corregido en SALIDA.md, segunda corrida: 1397832 bytes).

# SALIDA — ¿Go o Rust? (golandia 698), con el banco de la revancha

> Corrida del 2026-10-29 07:41. Regenerable con `python canales/golandia/recursos/pruebas-go-vs-rust/ejecutar_todo.py`. ⛔ Ninguna cifra de blog: todo salió de aquí.

- go: `go version go1.27.2 linux/amd64` (net/http, solo librería estándar)
- rust: `rustc 1.99.1 (58cf8f9c3 2026-10-13)` — axum 0.8.9 sobre tokio 1.54.1 (versiones fijadas con `=` en Cargo.toml y Cargo.lock)
- Docker `29.6.2`. Servidores con 2 CPU y 1g; `hey -c 16` con 4 CPU desde otro contenedor.
- otros contenedores encendidos al empezar (de otros agentes): 3 ['v694-tienda-web-1', 'v699-srv', 'v700-banco-127']

## 1 · Construir la imagen desde cero (`docker build --no-cache`, dependencias incluidas)

| lenguaje | qué se cronometra | segundos |
|---|---|---|
| go | go build | 5.9 |
| rust (glibc) | cargo build --release (baja y compila axum, tokio y 50 paquetes más) | 11.3 |
| rust (musl) | lo mismo, en rust:1.99.1-alpine | 20.9 |

## 2 · Lo que despliegas

| imagen | base | tamaño |
|---|---|---|
| go | scratch + binario estático | 2.4 MiB |
| rust (glibc) | debian:trixie-slim + binario (Rust 1.99.1, axum 0.8.9, tokio 1.54.1) | 29.0 MiB |
| rust (musl) | scratch + binario estático musl (Rust 1.99.1, axum 0.8.9, tokio 1.54.1) | 0.6 MiB |

## 0 · ¿Hacen lo mismo? (antes de medir)

- go: total 9990000, primeros ids [0, 1000, 2000, 3000, 4000]…, suma 99863963591 → igual que Go
- rust (glibc): total 9990000, primeros ids [0, 1000, 2000, 3000, 4000]…, suma 99863963591 → igual que Go
- rust (musl): total 9990000, primeros ids [0, 1000, 2000, 3000, 4000]…, suma 99863963591 → igual que Go

## 3 · Arranque: contenedor nuevo → primera respuesta 200 (mediana de 8) y memoria en reposo

| variante | 1ª respuesta (docker incluido) | «listo» (su propio reloj) | memoria en reposo |
|---|---|---|---|
| go | 226 ms | 0.0 ms | 2.0 MiB |
| rust (glibc) | 209 ms | 0.0 ms | 1.8 MiB |
| rust (musl) | 228 ms | 0.1 ms | 1.8 MiB |

## 4 · Carga con memoria: /trabajo (objetos, mapa, orden, JSON), a tope

`hey -c 16` desde otro contenedor; servidor con 2 CPU y 1g. «En frío»: los primeros 10 s tras arrancar. «Templado»: los 20 s siguientes.

| variante | peticiones/s en frío | peticiones/s templado | p50 | p99 | memoria máx. bajo carga |
|---|---|---|---|---|---|
| go | 454 | 440 | 29.3 ms | 126.9 ms | 32 MiB |
| rust (glibc) | 739 | 747 | 21.3 ms | 38.4 ms | 12 MiB |
| rust (musl) | 139 | 174 | 91.2 ms | 160.3 ms | 9 MiB |

## 5 · Carga de CPU pura: /suma, a tope

`hey -c 16` desde otro contenedor; servidor con 2 CPU y 1g. «En frío»: los primeros 10 s tras arrancar. «Templado»: los 20 s siguientes.

| variante | peticiones/s en frío | peticiones/s templado | p50 | p99 | memoria máx. bajo carga |
|---|---|---|---|---|---|
| go | 14258 | 14298 | 1.0 ms | 2.9 ms | 10 MiB |
| rust (glibc) | 15600 | 16155 | 1.0 ms | 1.7 ms | 5 MiB |
| rust (musl) | 15734 | 15829 | 1.0 ms | 1.7 ms | 3 MiB |

## 6 · Latencia de una petición sola: /trabajo

Un solo cliente a 5 peticiones/s (`hey -c 1 -q 5`), 20 s, tras 10 s de calentamiento a tope. Cada petición llega sola: no hay cola, es lo que tarda el servidor en atender una.

| variante | peticiones/s logradas | p50 | p99 |
|---|---|---|---|
| go | 5 | 3.90 ms | 7.70 ms |
| rust (glibc) | 5 | 3.10 ms | 6.50 ms |
| rust (musl) | 5 | 5.30 ms | 6.80 ms |

## 7 · Latencia de una petición sola: /suma

Un solo cliente a 5 peticiones/s (`hey -c 1 -q 5`), 20 s, tras 10 s de calentamiento a tope. Cada petición llega sola: no hay cola, es lo que tarda el servidor en atender una.

| variante | peticiones/s logradas | p50 | p99 |
|---|---|---|---|
| go | 5 | 0.50 ms | 1.20 ms |
| rust (glibc) | 5 | 0.40 ms | 1.10 ms |
| rust (musl) | 5 | 0.40 ms | 1.10 ms |

## 8 · Cómputo puro (reloj del propio proceso, sin red)

Como `investigacion/pruebas_go_vs_x_computo.py`: contenedor con 2 CPU y 1g, 50 peticiones de calentamiento y 300 medidas, de una en una; el tiempo es el campo «ns» que mide cada servidor alrededor de su cálculo.

| variante | /suma?n=200000 mediana | p99 | /trabajo?n=20000 mediana | p99 |
|---|---|---|---|---|
| go | 0.122 ms | 0.179 ms | 3.412 ms | 4.480 ms |
| rust (glibc) | 0.111 ms | 0.168 ms | 1.868 ms | 2.294 ms |
| rust (musl) | 0.111 ms | 0.154 ms | 3.970 ms | 4.915 ms |

## 9 · Lo que cuesta escribirlo

| | go | rust |
|---|---|---|
| líneas de código (sin blancos ni comentarios) | 90 | 67 |
| líneas totales del archivo | 102 | 79 |
| dependencias externas | 0 (solo librería estándar) | 2 directas (axum, tokio) · 51 paquetes en Cargo.lock |

Compilación dentro de `golang:1.27.2` y `rust:1.99.1-trixie` con `--cpus 2`, contenedor nuevo cada vez (cachés vacías), mediana de 3. «Incremental»: cambiar una cadena en el manejador de /espera y volver a compilar.

| paso | go (s) | rust (s) | las 3 corridas go · rust |
|---|---|---|---|
| bajar dependencias | — | 3.0 | — · 3.0 / 1.1 / 3.0 |
| compilación limpia (release) | 7.6 | 12.2 | 7.6 / 7.7 / 7.3 · 12.9 / 11.5 / 12.2 |
| incremental (release) | 0.14 | 1.39 | 0.2 / 0.1 / 0.1 · 1.4 / 1.3 / 1.4 |
| compilación limpia (debug) | (Go no tiene modo aparte) | 8.4 | — · 8.4 / 7.5 / 8.4 |
| incremental (debug) | — | 0.30 | — · 0.3 / 0.3 / 0.3 |

| binario | bytes | MiB |
|---|---|---|
| go (`-ldflags='-s -w'`) | 5873824 | 5.60 |
| rust glibc (`strip = true`) | 1282400 | 1.22 |
| rust musl (= la imagen scratch entera) | 660451 | 0.63 |

## Anexo · los núcleos que ve tokio dentro del contenedor

- `docker run --cpus 2 v698-govs-rust:1` → `LISTO_EN_MS 0.075 rust hilos=2`
- `docker run v698-govs-rust:1` → `LISTO_EN_MS 0.053 rust hilos=24`

`hilos` es `std::thread::available_parallelism()`, el número con el que `#[tokio::main]` crea sus hilos de trabajo.

- otros contenedores encendidos al terminar: 4
