# SALIDA — ¿Go o Python? (golandia), con el banco de la revancha

> Corrida del 2026-10-24 08:47. Regenerable con `python investigacion/pruebas_go_vs_x.py python`. ⛔ Ninguna cifra de blog: todo salió de aquí.

- go: `go version go1.27.2 linux/amd64` (net/http, solo librería estándar)
- python: `Python 3.13.17` — python:3.13.17-slim + FastAPI 0.162 + uvicorn 0.54
- Docker `29.6.2`. Servidores con 2 CPU y 1g; `hey -c 16` con 4 CPU desde otro contenedor.

## 1 · Construir la imagen desde cero (`docker build --no-cache`, dependencias incluidas)

| lenguaje | qué se cronometra | segundos |
|---|---|---|
| go | go build | 5.6 |
| python | pip install de FastAPI + uvicorn | 5.8 |

## 2 · Lo que despliegas

| imagen | base | tamaño |
|---|---|---|
| go | scratch + binario estático | 2.4 MiB |
| python | python:3.13.17-slim + FastAPI 0.162 + uvicorn 0.54 | 56.3 MiB |

## 0 · ¿Hacen lo mismo? (antes de medir)

- go: total 9990000, primeros ids [0, 1000, 2000, 3000, 4000]…, suma 99863963591 → igual que Go
- python (1 proceso): total 9990000, primeros ids [0, 1000, 2000, 3000, 4000]…, suma 99863963591 → igual que Go
- python (2 procesos): total 9990000, primeros ids [0, 1000, 2000, 3000, 4000]…, suma 99863963591 → igual que Go

## 3 · Arranque: contenedor nuevo → primera respuesta 200 (mediana de 8) y memoria en reposo

| variante | 1ª respuesta (docker incluido) | «listo» (su propio reloj) | memoria en reposo |
|---|---|---|---|
| go | 217 ms | 0.0 ms | 2.9 MiB |
| python (1 proceso) | 453 ms | 122.5 ms | 37.3 MiB |
| python (2 procesos) | 508 ms | 108.8 ms | 95.3 MiB |

## 4 · Carga con memoria: /trabajo (objetos, mapa, orden, JSON), a tope

`hey -c 16` desde otro contenedor; servidor con 2 CPU y 1g. «En frío»: los primeros 10 s tras arrancar. «Templado»: los 20 s siguientes.

| variante | peticiones/s en frío | peticiones/s templado | p50 | p99 | memoria máx. bajo carga |
|---|---|---|---|---|---|
| go | 475 | 475 | 26.8 ms | 122.1 ms | 38 MiB |
| python (1 proceso) | 73 | 75 | 198.8 ms | 482.8 ms | 101 MiB |
| python (2 procesos) | 148 | 149 | 87.5 ms | 293.8 ms | 168 MiB |

## 5 · Carga de CPU pura: /suma, a tope

`hey -c 16` desde otro contenedor; servidor con 2 CPU y 1g. «En frío»: los primeros 10 s tras arrancar. «Templado»: los 20 s siguientes.

| variante | peticiones/s en frío | peticiones/s templado | p50 | p99 | memoria máx. bajo carga |
|---|---|---|---|---|---|
| go | 14985 | 14926 | 1.0 ms | 2.6 ms | 11 MiB |
| python (1 proceso) | 110 | 111 | 135.8 ms | 328.7 ms | 41 MiB |
| python (2 procesos) | 233 | 234 | 64.7 ms | 148.0 ms | 101 MiB |

## 6 · Latencia de una petición sola: /trabajo

Un solo cliente a 5 peticiones/s (`hey -c 1 -q 5`), 20 s, tras 10 s de calentamiento a tope. Cada petición llega sola: no hay cola, es lo que tarda el servidor en atender una.

| variante | peticiones/s logradas | p50 | p99 |
|---|---|---|---|
| go | 5 | 4.00 ms | 5.10 ms |
| python (1 proceso) | 5 | 11.30 ms | 25.40 ms |
| python (2 procesos) | 5 | 11.30 ms | 27.50 ms |

## 7 · Latencia de una petición sola: /suma

Un solo cliente a 5 peticiones/s (`hey -c 1 -q 5`), 20 s, tras 10 s de calentamiento a tope. Cada petición llega sola: no hay cola, es lo que tarda el servidor en atender una.

| variante | peticiones/s logradas | p50 | p99 |
|---|---|---|---|
| go | 5 | 0.50 ms | 1.40 ms |
| python (1 proceso) | 5 | 9.20 ms | 10.70 ms |
| python (2 procesos) | 5 | 9.30 ms | 10.60 ms |

