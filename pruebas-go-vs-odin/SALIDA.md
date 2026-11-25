# SALIDA — ¿Go o Odin? (golandia 905), corrida 2, con el banco de la serie

> Corrida del 2026-11-23 13:02, con el candado `_logs/banco.lock` tomado. Regenerable con `python investigacion/pruebas_go_vs_x_2026_11_23.py odin 2` (mismo banco que 879 y 865). ⛔ Ninguna cifra de blog: todo salió de aquí.

- go: `go version go1.27.2 linux/amd64` (net/http, solo librería estándar)
- odin: `odin version dev-2026-11-nightly:9a69916 · odin-http 5e301be`
- Docker `29.6.2`. Servidores con 2 CPU y 1g; `hey -c 16` con 4 CPU desde otro contenedor.

## 1 · Construir la imagen desde cero (`docker build --no-cache`, dependencias incluidas)

| lenguaje | qué se cronometra | segundos |
|---|---|---|
| go | go build | 5.4 |
| odin | descarga de Odin dev-2026-11a + git clone de odin-http + odin build -o:speed | 52.6 |

## 2 · Lo que despliegas

| imagen | base | tamaño |
|---|---|---|
| go | scratch + binario estático | 2.4 MiB |
| odin | debian:trixie-slim + binario de odin build -o:speed (odin-http, io_uring) | 28.7 MiB |

El artefacto que se copia a la imagen: Go, 5.6 MiB; odin, 0.49 MiB.

## Estado de la máquina antes de medir

- /proc/loadavg (desde un contenedor alpine, VM de Docker con 24 hilos): `1.38 0.85 1.04 3/519 6`
- contenedores ajenos en marcha: ninguno

## 0 · ¿Hacen lo mismo? (antes de medir)

- go: total 9990000, primeros ids [0, 1000, 2000, 3000, 4000]…, suma 99863963591 → igual que Go
- odin (hilos por defecto): total 9990000, primeros ids [0, 1000, 2000, 3000, 4000]…, suma 99863963591 → igual que Go
- odin, HILOS=2: total 9990000, primeros ids [0, 1000, 2000, 3000, 4000]…, suma 99863963591 → igual que Go
- odin, HILOS=2, nombres con fmt: total 9990000, primeros ids [0, 1000, 2000, 3000, 4000]…, suma 99863963591 → igual que Go

## 3 · Arranque: contenedor nuevo → primera respuesta 200 (mediana de 8) y memoria en reposo

| variante | 1ª respuesta (docker incluido) | «listo» (su propio reloj) | memoria en reposo |
|---|---|---|---|
| go | 198 ms | 0.0 ms | 2.6 MiB |
| odin (hilos por defecto) | 213 ms | 0.0 ms | 12.4 MiB |
| odin, HILOS=2 | 206 ms | 0.0 ms | 2.4 MiB |
| odin, HILOS=2, nombres con fmt | 210 ms | 0.0 ms | 2.5 MiB |

## 4 · Carga con memoria: /trabajo (objetos, mapa, orden, JSON), a tope

`hey -c 16` desde otro contenedor; servidor con 2 CPU y 1g. «En frío»: los primeros 10 s tras arrancar. «Templado»: los 20 s siguientes.

| variante | peticiones/s en frío | peticiones/s templado | p50 | p99 | memoria máx. bajo carga |
|---|---|---|---|---|---|
| go | 475 | 455 | 28.0 ms | 123.8 ms | 31 MiB |
| odin (hilos por defecto) | 308 | 310 | 83.8 ms | 183.2 ms | 32 MiB |
| odin, HILOS=2 | 441 | 447 | 10.4 ms | 69.0 ms | 34 MiB |
| odin, HILOS=2, nombres con fmt | 367 | 366 | 43.3 ms | 52.2 ms | 9 MiB |

## 5 · Carga de CPU pura: /suma, a tope

`hey -c 16` desde otro contenedor; servidor con 2 CPU y 1g. «En frío»: los primeros 10 s tras arrancar. «Templado»: los 20 s siguientes.

| variante | peticiones/s en frío | peticiones/s templado | p50 | p99 | memoria máx. bajo carga |
|---|---|---|---|---|---|
| go | 14859 | 14981 | 1.0 ms | 2.6 ms | 10 MiB |
| odin (hilos por defecto) | 10472 | 9910 | 0.3 ms | 76.5 ms | 13 MiB |
| odin, HILOS=2 | 16187 | 16178 | 0.9 ms | 2.1 ms | 5 MiB |
| odin, HILOS=2, nombres con fmt | 16229 | 16131 | 0.9 ms | 2.7 ms | 5 MiB |

## 6 · Latencia de una petición sola: /trabajo

Un solo cliente a 5 peticiones/s (`hey -c 1 -q 5`), 20 s, tras 10 s de calentamiento a tope. Cada petición llega sola: no hay cola, es lo que tarda el servidor en atender una.

| variante | peticiones/s logradas | p50 | p99 |
|---|---|---|---|
| go | 5 | 4.00 ms | 5.20 ms |
| odin (hilos por defecto) | 5 | 5.10 ms | 6.10 ms |
| odin, HILOS=2 | 5 | 5.00 ms | 6.30 ms |
| odin, HILOS=2, nombres con fmt | 5 | 6.00 ms | 7.40 ms |

## 7 · Latencia de una petición sola: /suma

Un solo cliente a 5 peticiones/s (`hey -c 1 -q 5`), 20 s, tras 10 s de calentamiento a tope. Cada petición llega sola: no hay cola, es lo que tarda el servidor en atender una.

| variante | peticiones/s logradas | p50 | p99 |
|---|---|---|---|
| go | 5 | 0.50 ms | 2.40 ms |
| odin (hilos por defecto) | 5 | 0.50 ms | 1.70 ms |
| odin, HILOS=2 | 5 | 0.50 ms | 1.40 ms |
| odin, HILOS=2, nombres con fmt | 5 | 0.50 ms | 1.70 ms |

## 8 · Cómputo puro: el reloj del propio proceso, sin red

50 peticiones de calentamiento y 300 medidas, de una en una; el tiempo es el campo «ns» que mide cada servidor alrededor de su propio cálculo.

| variante | /suma?n=200000 mediana | p99 | /trabajo?n=20000 mediana | p99 |
|---|---|---|---|---|
| go | 0.122 ms | 0.161 ms | 3.332 ms | 3.820 ms |
| odin (hilos por defecto) | 0.113 ms | 0.152 ms | 4.568 ms | 5.090 ms |
| odin, HILOS=2 | 0.113 ms | 0.198 ms | 4.490 ms | 5.180 ms |
| odin, HILOS=2, nombres con fmt | 0.113 ms | 0.163 ms | 5.408 ms | 6.101 ms |

Al terminar: /proc/loadavg `1.49 1.61 1.38 1/552 6`; ajenos: ninguno

Vigilancia de renders (remotion / fab.py producir) cada 10 s durante toda la corrida: ninguno vivo (orquestar_902_905.py, 2026-11-23 13:14).
