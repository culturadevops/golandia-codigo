# SALIDA — ¿Go o Kotlin Native? (golandia 879), corrida 2, con el banco de la serie

> Corrida del 2026-11-18 09:57, con el candado `_logs/banco.lock` tomado. Regenerable con `python investigacion/pruebas_go_vs_nativo_2026_11.py kotlin-native 2` (mismo banco que 865). ⛔ Ninguna cifra de blog: todo salió de aquí.

- go: `go version go1.27.2 linux/amd64` (net/http, solo librería estándar)
- kotlin-native: `Kotlin 2.4.21 · Ktor 3.6.1 (CIO) · Gradle 9.9.1 (gradle:9.9.1-jdk21) · JVM: eclipse-temurin:25-jre · native-image 25.0.2 2026-01-20`
- Docker `29.6.2`. Servidores con 2 CPU y 1g; `hey -c 16` con 4 CPU desde otro contenedor.

## 1 · Construir la imagen desde cero (`docker build --no-cache`, dependencias incluidas)

| lenguaje | qué se cronometra | segundos |
|---|---|---|
| go | go build | 5.4 |
| kotlin JVM | gradle jarCompleto (Kotlin 2.4.21, Ktor 3.6.1 CIO) | 51.6 |
| kotlin GraalVM | gradle jarCompleto + native-image (GraalVM CE 25) | 62.6 |
| kotlin/native release | gradle linkReleaseExecutableLinuxX64 | 152.4 |
| kotlin/native debug | gradle linkDebugExecutableLinuxX64 | 143.8 |

## 2 · Lo que despliegas

| imagen | base | tamaño |
|---|---|---|
| go | scratch + binario estático | 2.4 MiB |
| kotlin-jvm | eclipse-temurin:25-jre + servidor.jar (Ktor y Kotlin dentro) | 121.5 MiB |
| kotlin-graal | debian:trixie-slim + binario de GraalVM native-image | 40.7 MiB |
| kotlin-native | debian:trixie-slim + binario de Kotlin/Native (release) | 29.6 MiB |
| kotlin-native-debug | debian:trixie-slim + binario de Kotlin/Native (debug) | 33.0 MiB |

El artefacto que se copia a la imagen: Go, 5.6 MiB; kotlin-jvm, 10.40 MiB; kotlin-graal, 34.94 MiB; kotlin-native, 3.73 MiB; kotlin-native-debug, 17.26 MiB.

## Estado de la máquina antes de medir

- /proc/loadavg (desde un contenedor alpine, VM de Docker con 24 hilos): `1.82 2.23 2.11 1/1137 6`
- contenedores ajenos en marcha (CPU en ese momento): jk-a-fijo 0.51%, jk-a-jenkins 0.61%, jk-b-jenkins-877 0.18%, jk-b-k8s-control-plane 5.43%

## 0 · ¿Hacen lo mismo? (antes de medir)

- go: total 9990000, primeros ids [0, 1000, 2000, 3000, 4000]…, suma 99863963591 → igual que Go
- kotlin JVM: total 9990000, primeros ids [0, 1000, 2000, 3000, 4000]…, suma 99863963591 → igual que Go
- kotlin GraalVM: total 9990000, primeros ids [0, 1000, 2000, 3000, 4000]…, suma 99863963591 → igual que Go
- kotlin/native: total 9990000, primeros ids [0, 1000, 2000, 3000, 4000]…, suma 99863963591 → igual que Go
- kotlin/native, HILOS=2: total 9990000, primeros ids [0, 1000, 2000, 3000, 4000]…, suma 99863963591 → igual que Go

## 3 · Arranque: contenedor nuevo → primera respuesta 200 (mediana de 8) y memoria en reposo

| variante | 1ª respuesta (docker incluido) | «listo» (su propio reloj) | memoria en reposo |
|---|---|---|---|
| go | 196 ms | 0.0 ms | 2.4 MiB |
| kotlin JVM | 454 ms | 195.1 ms | 55.2 MiB |
| kotlin GraalVM | 202 ms | 1.3 ms | 5.0 MiB |
| kotlin/native | 190 ms | 2.6 ms | 22.1 MiB |
| kotlin/native, HILOS=2 | 202 ms | 2.3 ms | 21.4 MiB |

## 4 · Carga con memoria: /trabajo (objetos, mapa, orden, JSON), a tope

`hey -c 16` desde otro contenedor; servidor con 2 CPU y 1g. «En frío»: los primeros 10 s tras arrancar. «Templado»: los 20 s siguientes.

| variante | peticiones/s en frío | peticiones/s templado | p50 | p99 | memoria máx. bajo carga |
|---|---|---|---|---|---|
| go | 475 | 475 | 27.1 ms | 119.4 ms | 32 MiB |
| kotlin JVM | 436 | 503 | 31.3 ms | 48.9 ms | 330 MiB |
| kotlin GraalVM | 400 | 413 | 39.3 ms | 47.8 ms | 289 MiB |
| kotlin/native | 108 | 109 | 145.3 ms | 246.5 ms | 169 MiB |
| kotlin/native, HILOS=2 | 120 | 120 | 133.3 ms | 141.6 ms | 94 MiB |

## 5 · Carga de CPU pura: /suma, a tope

`hey -c 16` desde otro contenedor; servidor con 2 CPU y 1g. «En frío»: los primeros 10 s tras arrancar. «Templado»: los 20 s siguientes.

| variante | peticiones/s en frío | peticiones/s templado | p50 | p99 | memoria máx. bajo carga |
|---|---|---|---|---|---|
| go | 14901 | 14945 | 1.0 ms | 2.6 ms | 12 MiB |
| kotlin JVM | 2411 | 4612 | 1.1 ms | 66.8 ms | 94 MiB |
| kotlin GraalVM | 3930 | 3809 | 1.4 ms | 66.0 ms | 37 MiB |
| kotlin/native | 4303 | 4329 | 1.6 ms | 55.7 ms | 205 MiB |
| kotlin/native, HILOS=2 | 4430 | 4409 | 1.6 ms | 55.9 ms | 196 MiB |

## 6 · Latencia de una petición sola: /trabajo

Un solo cliente a 5 peticiones/s (`hey -c 1 -q 5`), 20 s, tras 10 s de calentamiento a tope. Cada petición llega sola: no hay cola, es lo que tarda el servidor en atender una.

| variante | peticiones/s logradas | p50 | p99 |
|---|---|---|---|
| go | 5 | 4.60 ms | 8.40 ms |
| kotlin JVM | 5 | 4.50 ms | 7.60 ms |
| kotlin GraalVM | 5 | 4.70 ms | 10.80 ms |
| kotlin/native | 5 | 12.70 ms | 17.10 ms |
| kotlin/native, HILOS=2 | 5 | 11.90 ms | 15.20 ms |

## 7 · Latencia de una petición sola: /suma

Un solo cliente a 5 peticiones/s (`hey -c 1 -q 5`), 20 s, tras 10 s de calentamiento a tope. Cada petición llega sola: no hay cola, es lo que tarda el servidor en atender una.

| variante | peticiones/s logradas | p50 | p99 |
|---|---|---|---|
| go | 5 | 0.50 ms | 1.10 ms |
| kotlin JVM | 5 | 0.80 ms | 3.00 ms |
| kotlin GraalVM | 5 | 0.90 ms | 2.20 ms |
| kotlin/native | 5 | 0.90 ms | 2.50 ms |
| kotlin/native, HILOS=2 | 5 | 0.90 ms | 1.60 ms |

## 8 · Cómputo puro: el reloj del propio proceso, sin red

50 peticiones de calentamiento y 300 medidas, de una en una; el tiempo es el campo «ns» que mide cada servidor alrededor de su propio cálculo.

| variante | /suma?n=200000 mediana | p99 | /trabajo?n=20000 mediana | p99 |
|---|---|---|---|---|
| go | 0.122 ms | 0.154 ms | 3.276 ms | 3.761 ms |
| kotlin JVM | 0.101 ms | 0.154 ms | 2.650 ms | 5.028 ms |
| kotlin GraalVM | 0.134 ms | 0.174 ms | 3.900 ms | 8.848 ms |
| kotlin/native | 0.105 ms | 0.165 ms | 10.064 ms | 11.341 ms |
| kotlin/native, HILOS=2 | 0.105 ms | 0.149 ms | 10.009 ms | 11.498 ms |

Al terminar: /proc/loadavg `1.46 1.43 1.66 1/1135 6`; ajenos: jk-a-fijo 0.49%, jk-a-jenkins 0.56%, jk-b-jenkins-877 0.21%, jk-b-k8s-control-plane 5.57%

