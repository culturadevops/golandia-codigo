# SALIDA — ¿Go o Crystal? (golandia 902), corrida 2, con el banco de la serie

> Corrida del 2026-11-23 12:30, con el candado `_logs/banco.lock` tomado. Regenerable con `python investigacion/pruebas_go_vs_x_2026_11_23.py crystal 2` (mismo banco que 879 y 865). ⛔ Ninguna cifra de blog: todo salió de aquí.

- go: `go version go1.27.2 linux/amd64` (net/http, solo librería estándar)
- crystal: `Crystal 1.22.3 [3462c0774] (2026-11-13) ·  · LLVM: 22.1.3`
- Docker `29.6.2`. Servidores con 2 CPU y 1g; `hey -c 16` con 4 CPU desde otro contenedor.

## 1 · Construir la imagen desde cero (`docker build --no-cache`, dependencias incluidas)

| lenguaje | qué se cronometra | segundos |
|---|---|---|
| go | go build | 5.7 |
| crystal --release | crystal build --release --no-debug --static | 25.6 |
| crystal (modo por defecto) | crystal build --static (sin --release) | 12.6 |

## 2 · Lo que despliegas

| imagen | base | tamaño |
|---|---|---|
| go | scratch + binario estático | 2.4 MiB |
| crystal | scratch + binario estático (musl) de crystal build --release | 9.3 MiB |
| crystal-defecto | scratch + binario estático (musl) de crystal build, sin --release | 10.2 MiB |

El artefacto que se copia a la imagen: Go, 5.6 MiB; crystal, 21.90 MiB; crystal-defecto, 25.12 MiB.

## Estado de la máquina antes de medir

- /proc/loadavg (desde un contenedor alpine, VM de Docker con 24 hilos): `0.56 0.83 1.30 2/523 6`
- contenedores ajenos en marcha: ninguno

## 0 · ¿Hacen lo mismo? (antes de medir)

- go: total 9990000, primeros ids [0, 1000, 2000, 3000, 4000]…, suma 99863963591 → igual que Go
- crystal (1 hilo, por defecto): total 9990000, primeros ids [0, 1000, 2000, 3000, 4000]…, suma 99863963591 → igual que Go
- crystal, HILOS=2: total 9990000, primeros ids [0, 1000, 2000, 3000, 4000]…, suma 99863963591 → igual que Go
- crystal sin --release: total 9990000, primeros ids [0, 1000, 2000, 3000, 4000]…, suma 99863963591 → igual que Go

## 3 · Arranque: contenedor nuevo → primera respuesta 200 (mediana de 8) y memoria en reposo

| variante | 1ª respuesta (docker incluido) | «listo» (su propio reloj) | memoria en reposo |
|---|---|---|---|
| go | 207 ms | 0.0 ms | 2.9 MiB |
| crystal (1 hilo, por defecto) | 208 ms | 0.0 ms | 6.3 MiB |
| crystal, HILOS=2 | 195 ms | 0.0 ms | 6.4 MiB |
| crystal sin --release | 205 ms | 0.1 ms | 6.2 MiB |

## 4 · Carga con memoria: /trabajo (objetos, mapa, orden, JSON), a tope

`hey -c 16` desde otro contenedor; servidor con 2 CPU y 1g. «En frío»: los primeros 10 s tras arrancar. «Templado»: los 20 s siguientes.

| variante | peticiones/s en frío | peticiones/s templado | p50 | p99 | memoria máx. bajo carga |
|---|---|---|---|---|---|
| go | 470 | 475 | 26.8 ms | 119.4 ms | 30 MiB |
| crystal (1 hilo, por defecto) | 584 | 576 | 27.4 ms | 33.7 ms | 29 MiB |
| crystal, HILOS=2 | 126 | 130 | 122.7 ms | 242.8 ms | 24 MiB |
| crystal sin --release | 108 | 109 | 145.4 ms | 155.0 ms | 30 MiB |

## 5 · Carga de CPU pura: /suma, a tope

`hey -c 16` desde otro contenedor; servidor con 2 CPU y 1g. «En frío»: los primeros 10 s tras arrancar. «Templado»: los 20 s siguientes.

| variante | peticiones/s en frío | peticiones/s templado | p50 | p99 | memoria máx. bajo carga |
|---|---|---|---|---|---|
| go | 14971 | 14937 | 1.0 ms | 2.6 ms | 11 MiB |
| crystal (1 hilo, por defecto) | 7236 | 7250 | 2.2 ms | 2.7 ms | 8 MiB |
| crystal, HILOS=2 | 11634 | 11654 | 1.1 ms | 8.0 ms | 9 MiB |
| crystal sin --release | 2052 | 2052 | 7.7 ms | 8.6 ms | 8 MiB |

## 6 · Latencia de una petición sola: /trabajo

Un solo cliente a 5 peticiones/s (`hey -c 1 -q 5`), 20 s, tras 10 s de calentamiento a tope. Cada petición llega sola: no hay cola, es lo que tarda el servidor en atender una.

| variante | peticiones/s logradas | p50 | p99 |
|---|---|---|---|
| go | 5 | 7.00 ms | 9.80 ms |
| crystal (1 hilo, por defecto) | 5 | 5.40 ms | 9.80 ms |
| crystal, HILOS=2 | 5 | 4.00 ms | 7.30 ms |
| crystal sin --release | 5 | 11.80 ms | 16.00 ms |

## 7 · Latencia de una petición sola: /suma

Un solo cliente a 5 peticiones/s (`hey -c 1 -q 5`), 20 s, tras 10 s de calentamiento a tope. Cada petición llega sola: no hay cola, es lo que tarda el servidor en atender una.

| variante | peticiones/s logradas | p50 | p99 |
|---|---|---|---|
| go | 5 | 1.30 ms | 3.20 ms |
| crystal (1 hilo, por defecto) | 5 | 1.10 ms | 5.80 ms |
| crystal, HILOS=2 | 5 | 1.00 ms | 3.60 ms |
| crystal sin --release | 5 | 2.10 ms | 4.80 ms |

## 8 · Cómputo puro: el reloj del propio proceso, sin red

50 peticiones de calentamiento y 300 medidas, de una en una; el tiempo es el campo «ns» que mide cada servidor alrededor de su propio cálculo.

| variante | /suma?n=200000 mediana | p99 | /trabajo?n=20000 mediana | p99 |
|---|---|---|---|---|
| go | 0.122 ms | 0.173 ms | 3.251 ms | 3.552 ms |
| crystal (1 hilo, por defecto) | 0.132 ms | 0.179 ms | 1.627 ms | 2.493 ms |
| crystal, HILOS=2 | 0.132 ms | 0.205 ms | 1.601 ms | 2.190 ms |
| crystal sin --release | 0.483 ms | 0.635 ms | 9.064 ms | 9.893 ms |

Al terminar: /proc/loadavg `0.73 1.02 1.23 2/527 6`; ajenos: ninguno

Vigilancia de renders (remotion / fab.py producir) cada 10 s durante toda la corrida: ninguno vivo (orquestar_902_905.py, 2026-11-23 12:41).
