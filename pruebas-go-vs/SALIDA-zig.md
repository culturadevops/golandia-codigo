# SALIDA — Go contra zig, medidos en esta máquina

> Corrida del 2026-10-18 08:22. Regenerable con `python investigacion/pruebas_go_vs.py zig`.
> ⛔ Ninguna cifra de blog: todo salió de aquí. Mismo método que el vídeo 148 (Go contra Java).

- `Docker version 29.6.2, build dfc4efb`
- Go 1.27: `go version go1.27.2 linux/amd64` (compila en `golang:1.27.2`)
- Zig 0.17: `zig 0.17.0` (compila en `golandia-zig:0.17.0`)
- Trabajo de CPU por petición: 200.000 vueltas de multiplicar y módulo.

## 1 · Compilar (ms, imagen oficial, contenedor nuevo: caché vacía)

| lenguaje | corridas | mediana |
|---|---|---|
| Go 1.27 | [2943, 2925, 2920] | **2925 ms** |
| Zig 0.17 | [8915, 8952, 8895] | **8915 ms** |

## 2 · Lo que despliegas

| lenguaje | artefacto (/out) | imagen final (comprimida: lo que se descarga) |
|---|---|---|
| Go 1.27 | 6.594.720 B (6.29 MiB) | 2.71 MiB |
| Zig 0.17 | 4.062.376 B (3.87 MiB) | 1.04 MiB |

## 3 · Arranque: contenedor nuevo → primera respuesta (mediana de 8)

| lenguaje | 1ª respuesta (docker incluido) | proceso «listo» (su propio reloj) |
|---|---|---|
| Go 1.27 | 197 ms | 0.0 ms |
| Zig 0.17 | 191 ms | 0.0 ms |

Todas las corridas (externo, ms): go [224, 258, 195, 200, 189, 180, 200, 186] · zig [202, 197, 170, 180, 183, 185, 209, 200]

## 4 · Memoria (docker stats, MiB)

| lenguaje | reposo | tras 2.000 de CPU | bajo 300 concurrentes | tras la carga |
|---|---|---|---|---|
| Go 1.27 | 4.0 | 10.7 | 15.1 | 15.1 |
| Zig 0.17 | 1.6 | 2.4 | 78.1 | 78.1 |

## 5 · Templado (2.000 peticiones. n=200.000)

| lenguaje | ext p50 | ext p99 | cómputo primeras 20 | cómputo últimas 500 | suma devuelta |
|---|---|---|---|---|---|
| Go 1.27 | 1.11 ms | 26.51 ms | 0.122 ms | 0.122 ms | 99863963591 |
| Zig 0.17 | 0.98 ms | 26.25 ms | 0.112 ms | 0.109 ms | 99863963591 |

El «cómputo» es el campo `ns` que declara el propio servidor: el reloj del proceso, sin red ni Docker por medio.

## 6 · Concurrencia: 300 peticiones a la vez, 100 ms de espera cada una

| lenguaje | total | secuencial teórico |
|---|---|---|
| Go 1.27 | 2009 ms | 30.000 ms |
| Zig 0.17 | 2014 ms | 30.000 ms |

## 2b · Imagen en disco (docker images)

| imagen | en disco |
|---|---|
| golandia-govs-go:1 | 9.44MB |
| golandia-govs-zig:1 | 5.15MB |


## 1b · Segunda compilación (mismo contenedor, caché caliente, tras añadir una línea de comentario al fuente; mediana de 3)

| lenguaje | corridas | mediana |
|---|---|---|
| Go 1.27 | [174, 197, 184] | **184 ms** |
| Zig 0.17 | [7118, 7162, 7114] | **7118 ms** |

## 7 · Anexo: hilos del sistema bajo carga (docker stats, columna PIDS; 2026-10-18)

300 peticiones a la vez esperando 1.500 ms; se lee docker stats al segundo 1:

| lenguaje | reposo | bajo 300 a la vez |
|---|---|---|
| Go 1.27 | 5 hilos · 3,754 MiB | 14 hilos · 14,63 MiB |
| Zig 0.17 | 1 hilo · 1,516 MiB | 301 hilos · 91,61 MiB |

El servidor de Zig usa `std.Io` con su implementación por defecto (`Threaded`): cada tarea concurrente de `Group.concurrent`
corrió en un hilo del sistema. Go reparte 300 goroutines sobre 14 hilos.
