# SALIDA — Go contra elixir, medidos en esta máquina

> Corrida del 2026-10-18 08:23. Regenerable con `python investigacion/pruebas_go_vs.py elixir`.
> ⛔ Ninguna cifra de blog: todo salió de aquí. Mismo método que el vídeo 148 (Go contra Java).

- `Docker version 29.6.2, build dfc4efb`
- Go 1.27: `go version go1.27.2 linux/amd64` (compila en `golang:1.27.2`)
- Elixir 1.20 (OTP 29): `Elixir 1.20.4 (compiled with Erlang/OTP 29)` (compila en `golandia-elixir-compilar:1`)
- Trabajo de CPU por petición: 200.000 vueltas de multiplicar y módulo.

## 1 · Compilar (ms, imagen oficial, contenedor nuevo: caché vacía)

| lenguaje | corridas | mediana |
|---|---|---|
| Go 1.27 | [2921, 2973, 2957] | **2957 ms** |
| Elixir 1.20 (OTP 29) | [2085, 2071, 2084] | **2084 ms** |

## 2 · Lo que despliegas

| lenguaje | artefacto (/out) | imagen final (comprimida: lo que se descarga) |
|---|---|---|
| Go 1.27 | 6.594.720 B (6.29 MiB) | 2.71 MiB |
| Elixir 1.20 (OTP 29) | 67.637.667 B (64.50 MiB) | 58.16 MiB |

## 3 · Arranque: contenedor nuevo → primera respuesta (mediana de 8)

| lenguaje | 1ª respuesta (docker incluido) | proceso «listo» (su propio reloj) |
|---|---|---|
| Go 1.27 | 194 ms | 0.0 ms |
| Elixir 1.20 (OTP 29) | 465 ms | 255.2 ms |

Todas las corridas (externo, ms): go [184, 186, 202, 187, 209, 187, 232, 203] · elixir [466, 459, 448, 467, 454, 466, 464, 484]

## 4 · Memoria (docker stats, MiB)

| lenguaje | reposo | tras 2.000 de CPU | bajo 300 concurrentes | tras la carga |
|---|---|---|---|---|
| Go 1.27 | 3.1 | 10.2 | 15.1 | 14.8 |
| Elixir 1.20 (OTP 29) | 169.1 | 166.3 | 170.7 | 170.4 |

## 5 · Templado (2.000 peticiones. n=200.000)

| lenguaje | ext p50 | ext p99 | cómputo primeras 20 | cómputo últimas 500 | suma devuelta |
|---|---|---|---|---|---|
| Go 1.27 | 0.99 ms | 26.59 ms | 0.122 ms | 0.122 ms | 99863963591 |
| Elixir 1.20 (OTP 29) | 1.25 ms | 26.38 ms | 0.419 ms | 0.425 ms | 99863963591 |

El «cómputo» es el campo `ns` que declara el propio servidor: el reloj del proceso, sin red ni Docker por medio.

## 6 · Concurrencia: 300 peticiones a la vez, 100 ms de espera cada una

| lenguaje | total | secuencial teórico |
|---|---|---|
| Go 1.27 | 2012 ms | 30.000 ms |
| Elixir 1.20 (OTP 29) | 2013 ms | 30.000 ms |

Nota: en Elixir, «listo» se mide desde que arranca la máquina virtual (BEAM), no desde main: no hay un main equivalente.

## 2b · Imagen en disco (docker images)

| imagen | en disco |
|---|---|
| golandia-govs-go:1 | 9.44MB |
| golandia-govs-elixir:1 | 232MB |


## 1b · Segunda compilación (mismo contenedor, caché caliente, tras añadir una línea de comentario al fuente; mediana de 3)

| lenguaje | corridas | mediana |
|---|---|---|
| Go 1.27 | [174, 197, 184] | **184 ms** |
| Elixir 1.20 (OTP 29) | [356, 357, 348] | **356 ms** |

## 7 · Anexo: memoria de Elixir en reposo según los planificadores de la BEAM (docker stats a los 5 s; 2026-10-18)

| arranque | memoria en reposo | hilos del sistema |
|---|---|---|
| por defecto (un planificador por núcleo; la máquina tiene 24) | 150,5 MiB | 67 |
| ERL_FLAGS=+S 2:2 | 130,2 MiB | 23 |
| ERL_FLAGS=+S 1:1 | 125,3 MiB | 21 |

Menos planificadores bajan la memoria, pero no es lo principal: con uno solo sigue en ~125 MiB. No se averiguó el reparto del
resto (máquina virtual, código cargado, Bandit). La medición principal (169,1 MiB) se hizo con la configuración por defecto.
