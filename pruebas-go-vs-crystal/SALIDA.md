# SALIDA — ¿Go o Crystal? (golandia 902), corrida 1, con el banco de la serie

> Corrida del 2026-11-23 10:44, con el candado `_logs/banco.lock` tomado. Regenerable con `python investigacion/pruebas_go_vs_x_2026_11_23.py crystal 1` (mismo banco que 879 y 865). ⛔ Ninguna cifra de blog: todo salió de aquí.

- go: `go version go1.27.2 linux/amd64` (net/http, solo librería estándar)
- crystal: `Crystal 1.22.3 [3462c0774] (2026-11-13) ·  · LLVM: 22.1.3`
- Docker `29.6.2`. Servidores con 2 CPU y 1g; `hey -c 16` con 4 CPU desde otro contenedor.

## 1 · Construir la imagen desde cero (`docker build --no-cache`, dependencias incluidas)

| lenguaje | qué se cronometra | segundos |
|---|---|---|
| go | go build | 5.6 |
| crystal --release | crystal build --release --no-debug --static | 22.6 |
| crystal (modo por defecto) | crystal build --static (sin --release) | 11.3 |

## 2 · Lo que despliegas

| imagen | base | tamaño |
|---|---|---|
| go | scratch + binario estático | 2.4 MiB |
| crystal | scratch + binario estático (musl) de crystal build --release | 9.3 MiB |
| crystal-defecto | scratch + binario estático (musl) de crystal build, sin --release | 10.2 MiB |

El artefacto que se copia a la imagen: Go, 5.6 MiB; crystal, 21.90 MiB; crystal-defecto, 25.12 MiB.

## Estado de la máquina antes de medir

- /proc/loadavg (desde un contenedor alpine, VM de Docker con 24 hilos): `2.46 0.60 0.20 2/507 6`
- contenedores ajenos en marcha: ninguno

## 0 · ¿Hacen lo mismo? (antes de medir)

- go: total 9990000, primeros ids [0, 1000, 2000, 3000, 4000]…, suma 99863963591 → igual que Go
- crystal (1 hilo, por defecto): total 9990000, primeros ids [0, 1000, 2000, 3000, 4000]…, suma 99863963591 → igual que Go
- crystal, HILOS=2: total 9990000, primeros ids [0, 1000, 2000, 3000, 4000]…, suma 99863963591 → igual que Go
- crystal sin --release: total 9990000, primeros ids [0, 1000, 2000, 3000, 4000]…, suma 99863963591 → igual que Go

## 3 · Arranque: contenedor nuevo → primera respuesta 200 (mediana de 8) y memoria en reposo

| variante | 1ª respuesta (docker incluido) | «listo» (su propio reloj) | memoria en reposo |
|---|---|---|---|
| go | 196 ms | 0.0 ms | 2.6 MiB |
| crystal (1 hilo, por defecto) | 202 ms | 0.0 ms | 6.3 MiB |
| crystal, HILOS=2 | 199 ms | 0.0 ms | 6.1 MiB |
| crystal sin --release | 197 ms | 0.0 ms | 6.4 MiB |

## 4 · Carga con memoria: /trabajo (objetos, mapa, orden, JSON), a tope

`hey -c 16` desde otro contenedor; servidor con 2 CPU y 1g. «En frío»: los primeros 10 s tras arrancar. «Templado»: los 20 s siguientes.

| variante | peticiones/s en frío | peticiones/s templado | p50 | p99 | memoria máx. bajo carga |
|---|---|---|---|---|---|
| go | 481 | 475 | 26.6 ms | 122.3 ms | 30 MiB |
| crystal (1 hilo, por defecto) | 557 | 559 | 28.3 ms | 33.2 ms | 16 MiB |
| crystal, HILOS=2 | 137 | 110 | 144.4 ms | 269.4 ms | 24 MiB |
| crystal sin --release | 108 | 110 | 144.3 ms | 155.1 ms | 35 MiB |

## 5 · Carga de CPU pura: /suma, a tope

`hey -c 16` desde otro contenedor; servidor con 2 CPU y 1g. «En frío»: los primeros 10 s tras arrancar. «Templado»: los 20 s siguientes.

| variante | peticiones/s en frío | peticiones/s templado | p50 | p99 | memoria máx. bajo carga |
|---|---|---|---|---|---|
| go | 14981 | 15025 | 1.0 ms | 2.6 ms | 11 MiB |
| crystal (1 hilo, por defecto) | 7206 | 7233 | 2.2 ms | 2.7 ms | 7 MiB |
| crystal, HILOS=2 | 11659 | 11039 | 1.1 ms | 8.1 ms | 8 MiB |
| crystal sin --release | 2068 | 2065 | 7.7 ms | 8.7 ms | 8 MiB |

## 6 · Latencia de una petición sola: /trabajo

Un solo cliente a 5 peticiones/s (`hey -c 1 -q 5`), 20 s, tras 10 s de calentamiento a tope. Cada petición llega sola: no hay cola, es lo que tarda el servidor en atender una.

| variante | peticiones/s logradas | p50 | p99 |
|---|---|---|---|
| go | 5 | 6.70 ms | 10.40 ms |
| crystal (1 hilo, por defecto) | 5 | 2.80 ms | 5.00 ms |
| crystal, HILOS=2 | 5 | 2.10 ms | 5.50 ms |
| crystal sin --release | 5 | 9.50 ms | 11.30 ms |

## 7 · Latencia de una petición sola: /suma

Un solo cliente a 5 peticiones/s (`hey -c 1 -q 5`), 20 s, tras 10 s de calentamiento a tope. Cada petición llega sola: no hay cola, es lo que tarda el servidor en atender una.

| variante | peticiones/s logradas | p50 | p99 |
|---|---|---|---|
| go | 5 | 0.90 ms | 2.40 ms |
| crystal (1 hilo, por defecto) | 5 | 0.60 ms | 2.50 ms |
| crystal, HILOS=2 | 5 | 0.90 ms | 2.40 ms |
| crystal sin --release | 5 | 1.40 ms | 4.10 ms |

## 8 · Cómputo puro: el reloj del propio proceso, sin red

50 peticiones de calentamiento y 300 medidas, de una en una; el tiempo es el campo «ns» que mide cada servidor alrededor de su propio cálculo.

| variante | /suma?n=200000 mediana | p99 | /trabajo?n=20000 mediana | p99 |
|---|---|---|---|---|
| go | 0.122 ms | 0.158 ms | 3.269 ms | 3.840 ms |
| crystal (1 hilo, por defecto) | 0.132 ms | 0.164 ms | 1.647 ms | 2.510 ms |
| crystal, HILOS=2 | 0.132 ms | 0.167 ms | 1.675 ms | 2.513 ms |
| crystal sin --release | 0.485 ms | 0.577 ms | 9.031 ms | 9.969 ms |

Al terminar: /proc/loadavg `0.46 0.83 0.60 4/497 6`; ajenos: ninguno

Vigilancia de renders (remotion / fab.py producir) cada 10 s durante toda la corrida: ninguno vivo (orquestar_902_905.py, 2026-11-23 10:56).
