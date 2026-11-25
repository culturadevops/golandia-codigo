# SALIDA — Go contra csharp, medidos en esta máquina

> Corrida del 2026-10-18 08:12. Regenerable con `python investigacion/pruebas_go_vs.py csharp`.
> ⛔ Ninguna cifra de blog: todo salió de aquí. Mismo método que el vídeo 148 (Go contra Java).

- `Docker version 29.6.2, build dfc4efb`
- Go 1.27: `go version go1.27.2 linux/amd64` (compila en `golang:1.27.2`)
- C# .NET 10 (JIT): `10.0.402` (compila en `mcr.microsoft.com/dotnet/sdk:10.0`)
- C# .NET 10 (Native AOT): `10.0.402` (compila en `mcr.microsoft.com/dotnet/sdk:10.0-noble-aot`)
- Trabajo de CPU por petición: 200.000 vueltas de multiplicar y módulo.

## 1 · Compilar (ms, imagen oficial, contenedor nuevo: caché vacía)

| lenguaje | corridas | mediana |
|---|---|---|
| Go 1.27 | [3073, 2916, 2863] | **2916 ms** |
| C# .NET 10 (JIT) | [1839, 1892, 1892] | **1892 ms** |
| C# .NET 10 (Native AOT) | [14792, 13421, 14406] | **14406 ms** |

## 2 · Lo que despliegas

| lenguaje | artefacto (/out) | imagen final (comprimida: lo que se descarga) |
|---|---|---|
| Go 1.27 | 6.594.720 B (6.29 MiB) | 2.71 MiB |
| C# .NET 10 (JIT) | 109.850 B (0.10 MiB) | 91.54 MiB |
| C# .NET 10 (Native AOT) | 9.392.869 B (8.96 MiB) | 9.87 MiB |

## 3 · Arranque: contenedor nuevo → primera respuesta (mediana de 8)

| lenguaje | 1ª respuesta (docker incluido) | proceso «listo» (su propio reloj) |
|---|---|---|
| Go 1.27 | 193 ms | 0.0 ms |
| C# .NET 10 (JIT) | 352 ms | 115.5 ms |
| C# .NET 10 (Native AOT) | 493 ms | 300.9 ms |

Todas las corridas (externo, ms): go [348, 231, 186, 185, 200, 190, 195, 191] · csharp [354, 356, 388, 348, 349, 361, 350, 329] · csharp-aot [504, 496, 499, 483, 487, 489, 487, 500]

## 4 · Memoria (docker stats, MiB)

| lenguaje | reposo | tras 2.000 de CPU | bajo 300 concurrentes | tras la carga |
|---|---|---|---|---|
| Go 1.27 | 4.1 | 10.5 | 14.2 | 14.2 |
| C# .NET 10 (JIT) | 23.1 | 32.5 | 35.1 | 35.1 |
| C# .NET 10 (Native AOT) | 23.2 | 36.7 | 40.5 | 40.5 |

## 5 · Templado (2.000 peticiones. n=200.000)

| lenguaje | ext p50 | ext p99 | cómputo primeras 20 | cómputo últimas 500 | suma devuelta |
|---|---|---|---|---|---|
| Go 1.27 | 1.04 ms | 26.12 ms | 0.122 ms | 0.122 ms | 99863963591 |
| C# .NET 10 (JIT) | 1.34 ms | 27.07 ms | 0.136 ms | 0.119 ms | 99863963591 |
| C# .NET 10 (Native AOT) | 1.48 ms | 26.68 ms | 0.119 ms | 0.121 ms | 99863963591 |

El «cómputo» es el campo `ns` que declara el propio servidor: el reloj del proceso, sin red ni Docker por medio.

## 6 · Concurrencia: 300 peticiones a la vez, 100 ms de espera cada una

| lenguaje | total | secuencial teórico |
|---|---|---|
| Go 1.27 | 2013 ms | 30.000 ms |
| C# .NET 10 (JIT) | 2014 ms | 30.000 ms |
| C# .NET 10 (Native AOT) | 2011 ms | 30.000 ms |

## 7 · Anexo: el arranque de C# AOT depende del usuario (LISTO_EN_MS, 5 corridas)

La imagen chiseled de Microsoft corre por defecto sin privilegios (usuario 1654). Mismo binario, misma imagen:

| usuario | corridas (ms) |
|---|---|
| 1654 | 302.715 306.784 302.665 307.295 304.650 |
| root | 49.684 46.285 48.593 46.189 46.564 |

Probado y descartado como causa: HOME=/tmp, DOTNET_SYSTEM_NET_DISABLEIPV6=1, DOTNET_gcServer=0 (siguen en ~300 ms). La misma imagen con --user root baja a ~50 ms. No se averiguó qué hace .NET distinto sin privilegios.

## 2b · Imagen en disco (docker images)

| imagen | en disco |
|---|---|
| golandia-govs-go:1 | 9.44MB |
| golandia-govs-csharp:1 | 340MB |
| golandia-govs-csharp-aot:1 | 34.2MB |


## 1b · Segunda compilación (mismo contenedor, caché caliente, tras añadir una línea de comentario al fuente; mediana de 3)

| lenguaje | corridas | mediana |
|---|---|---|
| Go 1.27 | [174, 197, 184] | **184 ms** |
| C# .NET 10 (JIT) | [879, 844, 851] | **851 ms** |
| C# .NET 10 (Native AOT) | [8010, 8287, 8138] | **8138 ms** |
