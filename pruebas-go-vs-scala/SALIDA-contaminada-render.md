# SALIDA — ¿Go o Scala? (golandia 806), con el banco de la revancha

> ⛔ CONTAMINADA: desde las 09:48 un render (producir_y_subir.py) corrió a la vez, después de que el candado se borrara por huérfano. No se usa; se repite.

> Corrida del 2026-11-05 09:46, con el candado `_logs/banco.lock` tomado. Regenerable con `python investigacion/pruebas_go_vs_lote_2026_11.py scala` (mismo banco que 651-654 y 699). ⛔ Ninguna cifra de blog: todo salió de aquí.

- go: `go version go1.27.2 linux/amd64` (net/http, solo librería estándar)
- scala: `openjdk version "25.0.5" 2026-10-20 LTS · OpenJDK Runtime Environment Temurin-25.0.5+8 (build 25.0.5+8-LTS) · OpenJDK 64-Bit Server VM Temurin-25.0.5+8 (build 25.0.5+8-LTS, mixed mode, sharing)`
- Docker `29.6.2`. Servidores con 2 CPU y 1g; `hey -c 16` con 4 CPU desde otro contenedor.

## 1 · Construir la imagen desde cero (`docker build --no-cache`, dependencias incluidas)

| lenguaje | qué se cronometra | segundos |
|---|---|---|
| go | go build | 8.2 |
| scala | scala-cli package --assembly (Scala 3.10.0, descarga el compilador) | 37.0 |

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
| go | 232 ms | 0.0 ms | 3.0 MiB |
| scala (hilos virtuales) | 385 ms | 90.7 ms | 43.9 MiB |
| scala (hilos de plataforma) | 389 ms | 91.7 ms | 42.7 MiB |

## 4 · Carga con memoria: /trabajo (objetos, mapa, orden, JSON), a tope

`hey -c 16` desde otro contenedor; servidor con 2 CPU y 1g. «En frío»: los primeros 10 s tras arrancar. «Templado»: los 20 s siguientes.

| variante | peticiones/s en frío | peticiones/s templado | p50 | p99 | memoria máx. bajo carga |
|---|---|---|---|---|---|
| go | 413 | 444 | 29.0 ms | 130.7 ms | 33 MiB |
| scala (hilos virtuales) | 350 | 379 | 40.0 ms | 60.9 ms | 316 MiB |
| scala (hilos de plataforma) | 359 | 385 | 40.6 ms | 58.4 ms | 315 MiB |

## 5 · Carga de CPU pura: /suma, a tope

`hey -c 16` desde otro contenedor; servidor con 2 CPU y 1g. «En frío»: los primeros 10 s tras arrancar. «Templado»: los 20 s siguientes.

| variante | peticiones/s en frío | peticiones/s templado | p50 | p99 | memoria máx. bajo carga |
|---|---|---|---|---|---|
| go | 14413 | 14543 | 1.0 ms | 2.8 ms | 10 MiB |
| scala (hilos virtuales) | 13089 | 14656 | 0.6 ms | 15.8 ms | 65 MiB |
| scala (hilos de plataforma) | 12713 | 13722 | 1.0 ms | 6.4 ms | 63 MiB |

## 6 · Latencia de una petición sola: /trabajo

Un solo cliente a 5 peticiones/s (`hey -c 1 -q 5`), 20 s, tras 10 s de calentamiento a tope. Cada petición llega sola: no hay cola, es lo que tarda el servidor en atender una.

| variante | peticiones/s logradas | p50 | p99 |
|---|---|---|---|
| go | 5 | 4.10 ms | 5.70 ms |
| scala (hilos virtuales) | 5 | 5.40 ms | 7.50 ms |
| scala (hilos de plataforma) | 5 | 5.40 ms | 7.30 ms |

## 7 · Latencia de una petición sola: /suma

Un solo cliente a 5 peticiones/s (`hey -c 1 -q 5`), 20 s, tras 10 s de calentamiento a tope. Cada petición llega sola: no hay cola, es lo que tarda el servidor en atender una.

| variante | peticiones/s logradas | p50 | p99 |
|---|---|---|---|
| go | 5 | 0.50 ms | 1.60 ms |
| scala (hilos virtuales) | 5 | 0.60 ms | 1.90 ms |
| scala (hilos de plataforma) | 5 | 0.70 ms | 3.50 ms |

## 8 · Cómputo puro: el reloj del propio proceso, sin red

50 peticiones de calentamiento y 300 medidas, de una en una; el tiempo es el campo «ns» que mide cada servidor alrededor de su propio cálculo.

| variante | /suma?n=200000 mediana | p99 | /trabajo?n=20000 mediana | p99 |
|---|---|---|---|---|
| go | 0.123 ms | 0.255 ms | 3.548 ms | 5.122 ms |
| scala (hilos virtuales) | 0.105 ms | 0.164 ms | 4.801 ms | 14.339 ms |
| scala (hilos de plataforma) | 0.105 ms | 0.142 ms | 5.068 ms | 10.048 ms |

