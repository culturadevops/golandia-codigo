# SALIDA — ¿Go o Kotlin Native? (golandia 879), corrida 1, con el banco de la serie

> Corrida del 2026-11-18 09:37, con el candado `_logs/banco.lock` tomado. Regenerable con `python investigacion/pruebas_go_vs_nativo_2026_11.py kotlin-native 1` (mismo banco que 865). ⛔ Ninguna cifra de blog: todo salió de aquí.

- go: `go version go1.27.2 linux/amd64` (net/http, solo librería estándar)
- kotlin-native: `Kotlin 2.4.21 · Ktor 3.6.1 (CIO) · Gradle 9.9.1 (gradle:9.9.1-jdk21) · JVM: eclipse-temurin:25-jre · native-image 25.0.2 2026-01-20`
- Docker `29.6.2`. Servidores con 2 CPU y 1g; `hey -c 16` con 4 CPU desde otro contenedor.

## 1 · Construir la imagen desde cero (`docker build --no-cache`, dependencias incluidas)

| lenguaje | qué se cronometra | segundos |
|---|---|---|
| go | go build | 5.5 |
| kotlin JVM | gradle jarCompleto (Kotlin 2.4.21, Ktor 3.6.1 CIO) | 44.9 |
| kotlin GraalVM | gradle jarCompleto + native-image (GraalVM CE 25) | 67.0 |
| kotlin/native release | gradle linkReleaseExecutableLinuxX64 | 152.3 |
| kotlin/native debug | gradle linkDebugExecutableLinuxX64 | 136.9 |

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

- /proc/loadavg (desde un contenedor alpine, VM de Docker con 24 hilos): `1.80 2.59 1.93 1/1138 6`
- contenedores ajenos en marcha (CPU en ese momento): jk-a-fijo 0.59%, jk-a-jenkins 0.56%, jk-b-jenkins-877 0.25%, jk-b-k8s-control-plane 5.22%

## 0 · ¿Hacen lo mismo? (antes de medir)

- go: total 9990000, primeros ids [0, 1000, 2000, 3000, 4000]…, suma 99863963591 → igual que Go
- kotlin JVM: total 9990000, primeros ids [0, 1000, 2000, 3000, 4000]…, suma 99863963591 → igual que Go
- kotlin GraalVM: total 9990000, primeros ids [0, 1000, 2000, 3000, 4000]…, suma 99863963591 → igual que Go
- kotlin/native: total 9990000, primeros ids [0, 1000, 2000, 3000, 4000]…, suma 99863963591 → igual que Go
- kotlin/native, HILOS=2: total 9990000, primeros ids [0, 1000, 2000, 3000, 4000]…, suma 99863963591 → igual que Go

## 3 · Arranque: contenedor nuevo → primera respuesta 200 (mediana de 8) y memoria en reposo

| variante | 1ª respuesta (docker incluido) | «listo» (su propio reloj) | memoria en reposo |
|---|---|---|---|
| go | 194 ms | 0.0 ms | 2.4 MiB |
| kotlin JVM | 435 ms | 195.2 ms | 54.8 MiB |
| kotlin GraalVM | 208 ms | 1.3 ms | 4.8 MiB |
| kotlin/native | 195 ms | 2.3 ms | 21.6 MiB |
| kotlin/native, HILOS=2 | 194 ms | 2.3 ms | 22.0 MiB |

## 4 · Carga con memoria: /trabajo (objetos, mapa, orden, JSON), a tope

`hey -c 16` desde otro contenedor; servidor con 2 CPU y 1g. «En frío»: los primeros 10 s tras arrancar. «Templado»: los 20 s siguientes.

| variante | peticiones/s en frío | peticiones/s templado | p50 | p99 | memoria máx. bajo carga |
|---|---|---|---|---|---|
| go | 475 | 476 | 27.3 ms | 118.8 ms | 33 MiB |
| kotlin JVM | 429 | 501 | 31.3 ms | 52.8 ms | 332 MiB |
| kotlin GraalVM | 387 | 414 | 39.0 ms | 48.2 ms | 282 MiB |
| kotlin/native | 107 | 111 | 143.1 ms | 240.7 ms | 235 MiB |
| kotlin/native, HILOS=2 | 119 | 120 | 133.4 ms | 142.9 ms | 75 MiB |

## 5 · Carga de CPU pura: /suma, a tope

`hey -c 16` desde otro contenedor; servidor con 2 CPU y 1g. «En frío»: los primeros 10 s tras arrancar. «Templado»: los 20 s siguientes.

| variante | peticiones/s en frío | peticiones/s templado | p50 | p99 | memoria máx. bajo carga |
|---|---|---|---|---|---|
| go | 14900 | 14937 | 1.0 ms | 2.6 ms | 11 MiB |
| kotlin JVM | 2276 | 3014 | 1.2 ms | 76.3 ms | 104 MiB |
| kotlin GraalVM | 4108 | 3912 | 1.3 ms | 66.0 ms | 35 MiB |
| kotlin/native | 4387 | 4355 | 1.6 ms | 55.3 ms | 199 MiB |
| kotlin/native, HILOS=2 | 4517 | 4518 | 1.6 ms | 55.5 ms | 202 MiB |

## 6 · Latencia de una petición sola: /trabajo

Un solo cliente a 5 peticiones/s (`hey -c 1 -q 5`), 20 s, tras 10 s de calentamiento a tope. Cada petición llega sola: no hay cola, es lo que tarda el servidor en atender una.

| variante | peticiones/s logradas | p50 | p99 |
|---|---|---|---|
| go | 5 | 4.10 ms | 7.40 ms |
| kotlin JVM | 5 | 4.50 ms | 9.10 ms |
| kotlin GraalVM | 5 | 5.20 ms | 11.40 ms |
| kotlin/native | 5 | 12.30 ms | 15.10 ms |
| kotlin/native, HILOS=2 | 5 | 11.20 ms | 14.60 ms |

## 7 · Latencia de una petición sola: /suma

Un solo cliente a 5 peticiones/s (`hey -c 1 -q 5`), 20 s, tras 10 s de calentamiento a tope. Cada petición llega sola: no hay cola, es lo que tarda el servidor en atender una.

| variante | peticiones/s logradas | p50 | p99 |
|---|---|---|---|
| go | 5 | 0.60 ms | 1.60 ms |
| kotlin JVM | 5 | 0.90 ms | 3.10 ms |
| kotlin GraalVM | 5 | 0.70 ms | 1.60 ms |
| kotlin/native | 5 | 0.90 ms | 1.90 ms |
| kotlin/native, HILOS=2 | 5 | 0.90 ms | 1.90 ms |

## 8 · Cómputo puro: el reloj del propio proceso, sin red

50 peticiones de calentamiento y 300 medidas, de una en una; el tiempo es el campo «ns» que mide cada servidor alrededor de su propio cálculo.

| variante | /suma?n=200000 mediana | p99 | /trabajo?n=20000 mediana | p99 |
|---|---|---|---|---|
| go | 0.122 ms | 0.159 ms | 3.251 ms | 3.624 ms |
| kotlin JVM | 0.101 ms | 0.271 ms | 2.675 ms | 5.065 ms |
| kotlin GraalVM | 0.134 ms | 0.181 ms | 3.881 ms | 8.972 ms |
| kotlin/native | 0.111 ms | 0.143 ms | 10.062 ms | 11.294 ms |
| kotlin/native, HILOS=2 | 0.105 ms | 0.136 ms | 9.971 ms | 11.504 ms |

Al terminar: /proc/loadavg `1.43 2.01 1.90 2/1135 6`; ajenos: jk-a-fijo 0.50%, jk-a-jenkins 0.60%, jk-b-jenkins-877 0.14%, jk-b-k8s-control-plane 5.04%

