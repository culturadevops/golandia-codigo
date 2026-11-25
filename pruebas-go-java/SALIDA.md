# SALIDA — Go contra Java, medidos en esta máquina

> Corrida del 2026-08-29 08:21. Regenerable con
> `python investigacion/pruebas_go_java.py`. ⛔ Ninguna cifra de blog: todo salió de aquí.

- `Docker version 29.6.2, build dfc4efb`
- `go version go1.26.7 linux/amd64` (imagen oficial `golang:1.26`)
- `openjdk 21.0.12 2026-07-21 LTS` (imagen oficial `eclipse-temurin:21`)
- Mismo servidor HTTP, librería estándar en ambos; Java con hilos virtuales (Loom).
- Trabajo de CPU por petición: 200.000 vueltas de multiplicar y módulo.

## 1 · Compilar (ms, dentro de la imagen oficial; reloj del contenedor)

| lenguaje | corridas | mediana |
|---|---|---|
| go | [3291, 2975, 2965] | **2975 ms** |
| java | [305, 300, 297] | **300 ms** |

⚠️ Los 2.975 ms de Go son con la caché de compilación VACÍA (contenedor recién nacido): ahí Go
compila también la parte de la librería estándar que usas. En el día a día la caché existe.
Medido aparte (mismo contenedor, segunda compilación, `touch main.go` antes):

| lenguaje | segunda compilación |
|---|---|
| go (caché caliente) | **175 ms** |
| java (segunda javac) | **291 ms** |

## 2 · Lo que despliegas

| artefacto | bytes |
|---|---|
| go binario (-s -w) | 5.959.842 B (5.68 MiB) |
| go binario plano | 8.626.231 B (8.23 MiB) |
| java servidor.jar | 2.697 B (0.00 MiB) |

| imagen | descarga | en disco |
|---|---|---|
| `golandia-govsjava-go:1` | 2.42 MiB | 8.51MB |
| `golandia-govsjava-java:1` | 109.49 MiB | 454MB |
| `eclipse-temurin:21-jre` | 109.50 MiB | 459MB |
| `eclipse-temurin:21-jdk` | 213.39 MiB | 726MB |
| `golang:1.26` | 297.95 MiB | 1.29GB |

## 3 · Arranque: contenedor nuevo → primera respuesta (mediana de 8)

| lenguaje | 1ª respuesta (docker incluido) | proceso «listo» (su propio reloj) |
|---|---|---|
| go | 180 ms | 0.0 ms |
| java | 239 ms | 19.6 ms |

Todas las corridas (externo, ms): go [171, 192, 171, 166, 200, 189, 195, 166] · java [220, 247, 269, 231, 231, 247, 232, 263]

## 4 · Memoria (docker stats, MiB)

| lenguaje | reposo | tras 2.000 de CPU | bajo 300 concurrentes | tras la carga |
|---|---|---|---|---|
| go | 3.2 | 10.9 | 13.8 | 13.8 |
| java | 33.0 | 111.1 | 125.0 | 125.0 |

## 5 · Templado y el JIT (2.000 peticiones. n=200.000)

| lenguaje | ext p50 | ext p99 | cómputo primeras 20 | cómputo últimas 500 |
|---|---|---|---|---|
| go | 0.98 ms | 25.90 ms | 0.124 ms | 0.124 ms |
| java | 1.06 ms | 26.09 ms | 0.107 ms | 0.098 ms |

El «cómputo» es el campo `ns` que declara el propio servidor: el reloj del proceso,
sin red ni Docker por medio. Ahí se ve el JIT de Java trabajando.

## 6 · Concurrencia: 300 peticiones a la vez, 100 ms de espera cada una

| lenguaje | total | secuencial teórico |
|---|---|---|
| go | 2014 ms | 30.000 ms |
| java | 2004 ms | 30.000 ms |

Java va con el executor de hilos virtuales de Java 21. El Java de 2015 no está invitado.
