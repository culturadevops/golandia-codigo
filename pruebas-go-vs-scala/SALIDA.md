# SALIDA — ¿Go o Scala? (golandia 806), con el banco de la revancha

> Corrida del 2026-11-05 10:54, con el candado `_logs/banco.lock` tomado. Regenerable con `python investigacion/pruebas_go_vs_lote_2026_11.py scala` (mismo banco que 651-654 y 699). ⛔ Ninguna cifra de blog: todo salió de aquí.

- go: `go version go1.27.2 linux/amd64` (net/http, solo librería estándar)
- scala: `openjdk version "25.0.5" 2026-10-20 LTS · OpenJDK Runtime Environment Temurin-25.0.5+8 (build 25.0.5+8-LTS) · OpenJDK 64-Bit Server VM Temurin-25.0.5+8 (build 25.0.5+8-LTS, mixed mode, sharing)`
- Docker `29.6.2`. Servidores con 2 CPU y 1g; `hey -c 16` con 4 CPU desde otro contenedor.

## 1 · Construir la imagen desde cero (`docker build --no-cache`, dependencias incluidas)

| lenguaje | qué se cronometra | segundos |
|---|---|---|
| go | go build | 5.2 |
| scala | scala-cli package --assembly (Scala 3.10.0, descarga el compilador) | 26.6 |

## 2 · Lo que despliegas

| imagen | base | tamaño |
|---|---|---|
| go | scratch + binario estático | 2.4 MiB |
| scala | eclipse-temurin:25-jre + servidor.jar (con la librería de Scala dentro) | 120.7 MiB |

El artefacto que se copia a la imagen: Go, 5.6 MiB; scala, 9.40 MiB.

## 0 · ¿Hacen lo mismo? (antes de medir)

- go: total 9990000, primeros ids [0, 1000, 2000, 3000, 4000]…, suma 99863963591 → igual que Go
- scala (hilos virtuales): total 9990000, primeros ids [0, 1000, 2000, 3000, 4000]…, suma 99863963591 → igual que Go
- scala (hilos de plataforma): total 9990000, primeros ids [0, 1000, 2000, 3000, 4000]…, suma 99863963591 → igual que Go

## 3 · Arranque: contenedor nuevo → primera respuesta 200 (mediana de 8) y memoria en reposo

| variante | 1ª respuesta (docker incluido) | «listo» (su propio reloj) | memoria en reposo |
|---|---|---|---|
| go | 207 ms | 0.0 ms | 2.9 MiB |
| scala (hilos virtuales) | 325 ms | 80.9 ms | 43.5 MiB |
| scala (hilos de plataforma) | 334 ms | 81.9 ms | 43.0 MiB |

## 4 · Carga con memoria: /trabajo (objetos, mapa, orden, JSON), a tope

`hey -c 16` desde otro contenedor; servidor con 2 CPU y 1g. «En frío»: los primeros 10 s tras arrancar. «Templado»: los 20 s siguientes.

| variante | peticiones/s en frío | peticiones/s templado | p50 | p99 | memoria máx. bajo carga |
|---|---|---|---|---|---|
| go | 474 | 475 | 26.6 ms | 123.2 ms | 41 MiB |
| scala (hilos virtuales) | 368 | 408 | 38.0 ms | 51.0 ms | 314 MiB |
| scala (hilos de plataforma) | 364 | 399 | 39.3 ms | 55.2 ms | 314 MiB |

## 5 · Carga de CPU pura: /suma, a tope

`hey -c 16` desde otro contenedor; servidor con 2 CPU y 1g. «En frío»: los primeros 10 s tras arrancar. «Templado»: los 20 s siguientes.

| variante | peticiones/s en frío | peticiones/s templado | p50 | p99 | memoria máx. bajo carga |
|---|---|---|---|---|---|
| go | 14883 | 14842 | 1.0 ms | 2.7 ms | 11 MiB |
| scala (hilos virtuales) | 14466 | 15569 | 0.4 ms | 21.8 ms | 64 MiB |
| scala (hilos de plataforma) | 14354 | 15568 | 0.9 ms | 2.5 ms | 64 MiB |

## 6 · Latencia de una petición sola: /trabajo

Un solo cliente a 5 peticiones/s (`hey -c 1 -q 5`), 20 s, tras 10 s de calentamiento a tope. Cada petición llega sola: no hay cola, es lo que tarda el servidor en atender una.

| variante | peticiones/s logradas | p50 | p99 |
|---|---|---|---|
| go | 5 | 5.70 ms | 10.90 ms |
| scala (hilos virtuales) | 5 | 6.30 ms | 10.60 ms |
| scala (hilos de plataforma) | 5 | 6.40 ms | 12.20 ms |

## 7 · Latencia de una petición sola: /suma

Un solo cliente a 5 peticiones/s (`hey -c 1 -q 5`), 20 s, tras 10 s de calentamiento a tope. Cada petición llega sola: no hay cola, es lo que tarda el servidor en atender una.

| variante | peticiones/s logradas | p50 | p99 |
|---|---|---|---|
| go | 5 | 0.70 ms | 3.40 ms |
| scala (hilos virtuales) | 5 | 1.00 ms | 3.60 ms |
| scala (hilos de plataforma) | 5 | 1.10 ms | 3.50 ms |

## 8 · Cómputo puro: el reloj del propio proceso, sin red

50 peticiones de calentamiento y 300 medidas, de una en una; el tiempo es el campo «ns» que mide cada servidor alrededor de su propio cálculo.

| variante | /suma?n=200000 mediana | p99 | /trabajo?n=20000 mediana | p99 |
|---|---|---|---|---|
| go | 0.122 ms | 0.275 ms | 3.308 ms | 4.967 ms |
| scala (hilos virtuales) | 0.106 ms | 0.176 ms | 4.631 ms | 7.667 ms |
| scala (hilos de plataforma) | 0.106 ms | 0.171 ms | 4.516 ms | 7.608 ms |

