# SALIDA — ¿Go o Scala Native? (golandia 865), con el banco de la revancha

> Corrida del 2026-11-17 09:08, con el candado `_logs/banco.lock` tomado. Regenerable con `python investigacion/pruebas_go_vs_lote_2026_11.py scala-native` (mismo banco que 651-654 y 699). ⛔ Ninguna cifra de blog: todo salió de aquí.

- go: `go version go1.27.2 linux/amd64` (net/http, solo librería estándar)
- scala-native: `1.18.0 · Scala 3.10.0 · Scala Native 0.5.12 (Immix) · JVM: eclipse-temurin:25-jre`
- Docker `29.6.2`. Servidores con 2 CPU y 1g; `hey -c 16` con 4 CPU desde otro contenedor.

## 1 · Construir la imagen desde cero (`docker build --no-cache`, dependencias incluidas)

| lenguaje | qué se cronometra | segundos |
|---|---|---|
| go | go build | 5.6 |
| scala JVM | scala-cli package --assembly (Scala 3.10.0, Java 25) | 26.8 |
| scala native | scala-cli package --native --native-mode release-full | 24.5 |
| scala native + arreglo | release-full con SCALANATIVE_GC_TRAP_BASED_YIELDPOINTS=0 | 24.7 |
| scala native (modo por defecto) | scala-cli package --native (sin --native-mode) | 14.3 |

## 2 · Lo que despliegas

| imagen | base | tamaño |
|---|---|---|
| go | scratch + binario estático | 2.4 MiB |
| scala-jvm | eclipse-temurin:25-jre + servidor.jar | 120.7 MiB |
| scala-native | debian:trixie-slim + binario de Scala Native (release-full) | 30.1 MiB |
| scala-native-sintrap | debian:trixie-slim + binario release-full con el arreglo | 30.1 MiB |
| scala-native-defecto | debian:trixie-slim + binario de Scala Native (modo por defecto) | 30.1 MiB |

El artefacto que se copia a la imagen: Go, 5.6 MiB; scala-jvm, 9.40 MiB; scala-native, 5.49 MiB; scala-native-sintrap, 5.44 MiB; scala-native-defecto, 6.09 MiB.

## 0 · ¿Hacen lo mismo? (antes de medir)

- go: total 9990000, primeros ids [0, 1000, 2000, 3000, 4000]…, suma 99863963591 → igual que Go
- scala JVM: total 9990000, primeros ids [0, 1000, 2000, 3000, 4000]…, suma 99863963591 → igual que Go
- scala native: total 9990000, primeros ids [0, 1000, 2000, 3000, 4000]…, suma 99863963591 → igual que Go
- scala native, HILOS=2: total 9990000, primeros ids [0, 1000, 2000, 3000, 4000]…, suma 99863963591 → igual que Go

## 3 · Arranque: contenedor nuevo → primera respuesta 200 (mediana de 8) y memoria en reposo

| variante | 1ª respuesta (docker incluido) | «listo» (su propio reloj) | memoria en reposo |
|---|---|---|---|
| go | 186 ms | 0.0 ms | 2.8 MiB |
| scala JVM | 294 ms | 69.2 ms | 39.9 MiB |
| scala native | 201 ms | 0.1 ms | 3.2 MiB |
| scala native, HILOS=2 | 198 ms | 0.1 ms | 3.2 MiB |

## 4 · Carga con memoria: /trabajo (objetos, mapa, orden, JSON), a tope

`hey -c 16` desde otro contenedor; servidor con 2 CPU y 1g. «En frío»: los primeros 10 s tras arrancar. «Templado»: los 20 s siguientes.

| variante | peticiones/s en frío | peticiones/s templado | p50 | p99 | memoria máx. bajo carga |
|---|---|---|---|---|---|
| go | 480 | 480 | 26.9 ms | 122.2 ms | 34 MiB |
| scala JVM | 369 | 393 | 40.6 ms | 44.8 ms | 297 MiB |
| scala native | 35 | 36 | 447.3 ms | 783.4 ms | 802 MiB |
| scala native, HILOS=2 | 96 | 100 | 156.5 ms | 195.5 ms | 171 MiB |

## 5 · Carga de CPU pura: /suma, a tope

`hey -c 16` desde otro contenedor; servidor con 2 CPU y 1g. «En frío»: los primeros 10 s tras arrancar. «Templado»: los 20 s siguientes.

| variante | peticiones/s en frío | peticiones/s templado | p50 | p99 | memoria máx. bajo carga |
|---|---|---|---|---|---|
| go | 15075 | 15081 | 1.0 ms | 2.6 ms | 11 MiB |
| scala JVM | 14284 | 14730 | 1.0 ms | 4.7 ms | 55 MiB |
| scala native | 5905 | 5755 | 0.4 ms | 83.2 ms | 28 MiB |
| scala native, HILOS=2 | 10899 | 10878 | 1.3 ms | 8.6 ms | 28 MiB |

## 6 · Latencia de una petición sola: /trabajo

Un solo cliente a 5 peticiones/s (`hey -c 1 -q 5`), 20 s, tras 10 s de calentamiento a tope. Cada petición llega sola: no hay cola, es lo que tarda el servidor en atender una.

| variante | peticiones/s logradas | p50 | p99 |
|---|---|---|---|
| go | 5 | 7.10 ms | 10.00 ms |
| scala JVM | 5 | 7.80 ms | 11.20 ms |
| scala native | 5 | 17.60 ms | 24.20 ms |
| scala native, HILOS=2 | 5 | 19.40 ms | 23.60 ms |

## 7 · Latencia de una petición sola: /suma

Un solo cliente a 5 peticiones/s (`hey -c 1 -q 5`), 20 s, tras 10 s de calentamiento a tope. Cada petición llega sola: no hay cola, es lo que tarda el servidor en atender una.

| variante | peticiones/s logradas | p50 | p99 |
|---|---|---|---|
| go | 5 | 1.30 ms | 3.70 ms |
| scala JVM | 5 | 1.10 ms | 3.70 ms |
| scala native | 5 | 1.10 ms | 6.90 ms |
| scala native, HILOS=2 | 5 | 1.20 ms | 5.30 ms |

## 8 · Cómputo puro: el reloj del propio proceso, sin red

50 peticiones de calentamiento y 300 medidas, de una en una; el tiempo es el campo «ns» que mide cada servidor alrededor de su propio cálculo.

| variante | /suma?n=200000 mediana | p99 | /trabajo?n=20000 mediana | p99 |
|---|---|---|---|---|
| go | 0.122 ms | 0.211 ms | 3.233 ms | 3.611 ms |
| scala JVM | 0.105 ms | 0.167 ms | 4.569 ms | 7.058 ms |
| scala native | 0.149 ms | 0.223 ms | 18.872 ms | 21.653 ms |
| scala native, HILOS=2 | 0.149 ms | 0.211 ms | 18.861 ms | 21.137 ms |

