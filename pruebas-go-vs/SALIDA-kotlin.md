# SALIDA — Go contra kotlin, medidos en esta máquina

> Corrida del 2026-10-18 08:20. Regenerable con `python investigacion/pruebas_go_vs.py kotlin`.
> ⛔ Ninguna cifra de blog: todo salió de aquí. Mismo método que el vídeo 148 (Go contra Java).

- `Docker version 29.6.2, build dfc4efb`
- Go 1.27: `go version go1.27.2 linux/amd64` (compila en `golang:1.27.2`)
- Kotlin 2.4 (JVM 21): `info: kotlinc-jvm 2.4.21 (JRE 21.0.12+8-LTS)` (compila en `golandia-kotlinc:2.4.21`)
- Trabajo de CPU por petición: 200.000 vueltas de multiplicar y módulo.

## 1 · Compilar (ms, imagen oficial, contenedor nuevo: caché vacía)

| lenguaje | corridas | mediana |
|---|---|---|
| Go 1.27 | [2914, 2946, 2979] | **2946 ms** |
| Kotlin 2.4 (JVM 21) | [2664, 2669, 2601] | **2664 ms** |

## 2 · Lo que despliegas

| lenguaje | artefacto (/out) | imagen final (comprimida: lo que se descarga) |
|---|---|---|
| Go 1.27 | 6.594.720 B (6.29 MiB) | 2.71 MiB |
| Kotlin 2.4 (JVM 21) | 5.682.626 B (5.42 MiB) | 114.32 MiB |

## 3 · Arranque: contenedor nuevo → primera respuesta (mediana de 8)

| lenguaje | 1ª respuesta (docker incluido) | proceso «listo» (su propio reloj) |
|---|---|---|
| Go 1.27 | 187 ms | 0.0 ms |
| Kotlin 2.4 (JVM 21) | 249 ms | 20.2 ms |

Todas las corridas (externo, ms): go [229, 168, 178, 203, 185, 174, 211, 189] · kotlin [307, 230, 247, 280, 252, 240, 241, 270]

## 4 · Memoria (docker stats, MiB)

| lenguaje | reposo | tras 2.000 de CPU | bajo 300 concurrentes | tras la carga |
|---|---|---|---|---|
| Go 1.27 | 3.5 | 10.5 | 14.7 | 14.7 |
| Kotlin 2.4 (JVM 21) | 36.5 | 112.9 | 115.6 | 115.3 |

## 5 · Templado (2.000 peticiones. n=200.000)

| lenguaje | ext p50 | ext p99 | cómputo primeras 20 | cómputo últimas 500 | suma devuelta |
|---|---|---|---|---|---|
| Go 1.27 | 1.16 ms | 26.43 ms | 0.122 ms | 0.122 ms | 99863963591 |
| Kotlin 2.4 (JVM 21) | 1.19 ms | 27.21 ms | 0.109 ms | 0.100 ms | 99863963591 |

El «cómputo» es el campo `ns` que declara el propio servidor: el reloj del proceso, sin red ni Docker por medio.

## 6 · Concurrencia: 300 peticiones a la vez, 100 ms de espera cada una

| lenguaje | total | secuencial teórico |
|---|---|---|
| Go 1.27 | 2009 ms | 30.000 ms |
| Kotlin 2.4 (JVM 21) | 2011 ms | 30.000 ms |

## 2b · Imagen en disco (docker images)

| imagen | en disco |
|---|---|
| golandia-govs-go:1 | 9.44MB |
| golandia-govs-kotlin:1 | 465MB |


## 1b · Segunda compilación (mismo contenedor, caché caliente, tras añadir una línea de comentario al fuente; mediana de 3)

| lenguaje | corridas | mediana |
|---|---|---|
| Go 1.27 | [174, 197, 184] | **184 ms** |
| Kotlin 2.4 (JVM 21) | [2657, 2667, 2676] | **2667 ms** |
