# SALIDA — ¿Go o Odin? (golandia 905), corrida 1, con el banco de la serie

> Corrida del 2026-11-23 12:14, con el candado `_logs/banco.lock` tomado. Regenerable con `python investigacion/pruebas_go_vs_x_2026_11_23.py odin 1` (mismo banco que 879 y 865). ⛔ Ninguna cifra de blog: todo salió de aquí.

- go: `go version go1.27.2 linux/amd64` (net/http, solo librería estándar)
- odin: `odin version dev-2026-11-nightly:9a69916 · odin-http 5e301be`
- Docker `29.6.2`. Servidores con 2 CPU y 1g; `hey -c 16` con 4 CPU desde otro contenedor.

## 1 · Construir la imagen desde cero (`docker build --no-cache`, dependencias incluidas)

| lenguaje | qué se cronometra | segundos |
|---|---|---|
| go | go build | 5.6 |
| odin | descarga de Odin dev-2026-11a + git clone de odin-http + odin build -o:speed | 55.2 |

## 2 · Lo que despliegas

| imagen | base | tamaño |
|---|---|---|
| go | scratch + binario estático | 2.4 MiB |
| odin | debian:trixie-slim + binario de odin build -o:speed (odin-http, io_uring) | 28.7 MiB |

El artefacto que se copia a la imagen: Go, 5.6 MiB; odin, 0.48 MiB.

## Estado de la máquina antes de medir

- /proc/loadavg (desde un contenedor alpine, VM de Docker con 24 hilos): `2.25 1.98 1.83 1/519 6`
- contenedores ajenos en marcha: ninguno

## 0 · ¿Hacen lo mismo? (antes de medir)

- go: total 9990000, primeros ids [0, 1000, 2000, 3000, 4000]…, suma 99863963591 → igual que Go
- odin (hilos por defecto): total 9990000, primeros ids [0, 1000, 2000, 3000, 4000]…, suma 99863963591 → igual que Go
- odin, HILOS=2: total 9990000, primeros ids [0, 1000, 2000, 3000, 4000]…, suma 99863963591 → igual que Go

## 3 · Arranque: contenedor nuevo → primera respuesta 200 (mediana de 8) y memoria en reposo

| variante | 1ª respuesta (docker incluido) | «listo» (su propio reloj) | memoria en reposo |
|---|---|---|---|
| go | 190 ms | 0.0 ms | 2.8 MiB |
| odin (hilos por defecto) | 199 ms | 0.0 ms | 12.5 MiB |
| odin, HILOS=2 | 206 ms | 0.0 ms | 2.5 MiB |

## 4 · Carga con memoria: /trabajo (objetos, mapa, orden, JSON), a tope

`hey -c 16` desde otro contenedor; servidor con 2 CPU y 1g. «En frío»: los primeros 10 s tras arrancar. «Templado»: los 20 s siguientes.

| variante | peticiones/s en frío | peticiones/s templado | p50 | p99 | memoria máx. bajo carga |
|---|---|---|---|---|---|
| go | 475 | 475 | 26.8 ms | 122.2 ms | 32 MiB |
| odin (hilos por defecto) | 323 | 237 | 92.2 ms | 196.8 ms | 42 MiB |
| odin, HILOS=2 | 368 | 363 | 43.4 ms | 55.8 ms | 9 MiB |

## 5 · Carga de CPU pura: /suma, a tope

`hey -c 16` desde otro contenedor; servidor con 2 CPU y 1g. «En frío»: los primeros 10 s tras arrancar. «Templado»: los 20 s siguientes.

| variante | peticiones/s en frío | peticiones/s templado | p50 | p99 | memoria máx. bajo carga |
|---|---|---|---|---|---|
| go | 14939 | 14938 | 1.0 ms | 2.6 ms | 10 MiB |
| odin (hilos por defecto) | 10819 | 9157 | 0.3 ms | 76.4 ms | 14 MiB |
| odin, HILOS=2 | 15888 | 15561 | 1.0 ms | 2.1 ms | 4 MiB |

## 6 · Latencia de una petición sola: /trabajo

Un solo cliente a 5 peticiones/s (`hey -c 1 -q 5`), 20 s, tras 10 s de calentamiento a tope. Cada petición llega sola: no hay cola, es lo que tarda el servidor en atender una.

| variante | peticiones/s logradas | p50 | p99 |
|---|---|---|---|
| go | 5 | 4.10 ms | 5.00 ms |
| odin (hilos por defecto) | 5 | 6.00 ms | 7.30 ms |
| odin, HILOS=2 | 5 | 5.90 ms | 7.30 ms |

## 7 · Latencia de una petición sola: /suma

Un solo cliente a 5 peticiones/s (`hey -c 1 -q 5`), 20 s, tras 10 s de calentamiento a tope. Cada petición llega sola: no hay cola, es lo que tarda el servidor en atender una.

| variante | peticiones/s logradas | p50 | p99 |
|---|---|---|---|
| go | 5 | 0.50 ms | 1.40 ms |
| odin (hilos por defecto) | 5 | 0.50 ms | 1.30 ms |
| odin, HILOS=2 | 5 | 0.50 ms | 1.40 ms |

## 8 · Cómputo puro: el reloj del propio proceso, sin red

50 peticiones de calentamiento y 300 medidas, de una en una; el tiempo es el campo «ns» que mide cada servidor alrededor de su propio cálculo.

| variante | /suma?n=200000 mediana | p99 | /trabajo?n=20000 mediana | p99 |
|---|---|---|---|---|
| go | 0.122 ms | 0.180 ms | 3.363 ms | 4.104 ms |
| odin (hilos por defecto) | 0.117 ms | 0.174 ms | 5.514 ms | 6.185 ms |
| odin, HILOS=2 | 0.117 ms | 0.149 ms | 5.446 ms | 5.961 ms |

Al terminar: /proc/loadavg `1.46 1.73 1.74 4/550 6`; ajenos: ninguno

Vigilancia de renders (remotion / fab.py producir) cada 10 s durante toda la corrida: ninguno vivo (orquestar_902_905.py, 2026-11-23 12:23).
