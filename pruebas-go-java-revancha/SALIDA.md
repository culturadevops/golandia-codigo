# SALIDA — la revancha de Go contra Java (vídeo 148), medida en esta máquina

> Corrida del 2026-10-19 15:57. Regenerable con `python investigacion/pruebas_go_java_revancha.py`. ⛔ Ninguna cifra de blog: todo salió de aquí.

- docker: `29.6.2`
- go: `go version go1.26.9 linux/amd64`
- java: `openjdk version "25.0.4.1" 2026-08-18 LTS`
- graal: `native-image 25.0.2 2026-01-20`
- Máquina: 24 hilos (Windows 11 + Docker Desktop). Servidores limitados a 2 CPU y 1g; `hey` con 4 CPU.
- /trabajo?n=20000: crea 20000 objetos, los mete en un mapa, los ordena y arma un JSON a mano (igual en los dos).
- /suma?n=200000: el trabajo de CPU del vídeo 148 (multiplicar y módulo).

## 1 · Compilar (dentro de la imagen oficial de cada uno, `time -p`, caché vacía)

| paso | imagen | segundos |
|---|---|---|
| go build | `golang:1.26` | 3.0 |
| javac + jar | `eclipse-temurin:25-jdk` | 0.4 |
| jlink (JRE mínimo) | `eclipse-temurin:25-jdk` | 1.4 |
| native-image (GraalVM) | `ghcr.io/graalvm/native-image-community:25` | 13.3 |

## 2 · Lo que despliegas

| imagen | base | descarga (comprimida) |
|---|---|---|
| go | scratch | 2.4 MiB |
| java-jre | eclipse-temurin:25-jre | 112.0 MiB |
| java-jlink | debian:trixie-slim + JRE de jlink | 50.3 MiB |
| java-nativo | debian:trixie-slim + binario nativo | 33.6 MiB |

El JRE de jlink (java.base + jdk.httpserver) pesa 42.0 MiB; el binario de GraalVM, 14.6 MiB; el de Go, 5.5 MiB.

## 0 · ¿Hacen lo mismo? (antes de medir)

- go: total 9990000, primeros ids [0, 1000, 2000, 3000, 4000]…, suma 99863963591 → igual que Go
- java-jre (virtuales): total 9990000, primeros ids [0, 1000, 2000, 3000, 4000]…, suma 99863963591 → igual que Go
- java-jre (plataforma): total 9990000, primeros ids [0, 1000, 2000, 3000, 4000]…, suma 99863963591 → igual que Go
- java-jlink: total 9990000, primeros ids [0, 1000, 2000, 3000, 4000]…, suma 99863963591 → igual que Go
- java-nativo (GraalVM): total 9990000, primeros ids [0, 1000, 2000, 3000, 4000]…, suma 99863963591 → igual que Go

## 3 · Arranque: contenedor nuevo → primera respuesta 200 (mediana de 8) y memoria en reposo

| variante | 1ª respuesta (docker incluido) | «listo» (su propio reloj) | memoria en reposo |
|---|---|---|---|
| go | 242 ms | 0.0 ms | 3.2 MiB |
| java-jre (virtuales) | 310 ms | 19.3 ms | 34.8 MiB |
| java-jre (plataforma) | 304 ms | 20.1 ms | 34.2 MiB |
| java-jlink | 362 ms | 48.9 ms | 36.3 MiB |
| java-nativo (GraalVM) | 258 ms | 0.6 ms | 5.7 MiB |

## 4 · Carga con memoria: /trabajo (objetos, mapa, orden, JSON)

`hey -c 16` desde otro contenedor; servidor con 2 CPU y 1g de memoria. «En frío»: los primeros 10 s tras arrancar. «Templado»: los 20 s siguientes.

| variante | peticiones/s en frío | peticiones/s templado | p50 | p99 | memoria máx. bajo carga |
|---|---|---|---|---|---|
| go | 353 | 385 | 32.9 ms | 157.8 ms | 36 MiB |
| java-jre (virtuales) | 885 | 1012 | 15.5 ms | 21.6 ms | 296 MiB |
| java-jre (plataforma) | 908 | 1003 | 15.9 ms | 21.3 ms | 292 MiB |
| java-jlink | 943 | 1034 | 15.4 ms | 19.2 ms | 293 MiB |
| java-nativo (GraalVM) | 653 | 712 | 20.6 ms | 33.5 ms | 275 MiB |

## 5 · Carga de CPU pura: /suma (aquí se ve si los hilos virtuales molestan)

`hey -c 16` desde otro contenedor; servidor con 2 CPU y 1g de memoria. «En frío»: los primeros 10 s tras arrancar. «Templado»: los 20 s siguientes.

| variante | peticiones/s en frío | peticiones/s templado | p50 | p99 | memoria máx. bajo carga |
|---|---|---|---|---|---|
| go | 14241 | 14333 | 1.0 ms | 2.8 ms | 9 MiB |
| java-jre (virtuales) | 13335 | 14217 | 0.4 ms | 21.3 ms | 54 MiB |
| java-jre (plataforma) | 13808 | 14495 | 1.0 ms | 2.8 ms | 56 MiB |
| java-jlink | 13704 | 14535 | 0.4 ms | 22.1 ms | 60 MiB |
| java-nativo (GraalVM) | 10376 | 10257 | 1.4 ms | 5.0 ms | 10 MiB |

