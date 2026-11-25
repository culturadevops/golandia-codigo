# SALIDA — ¿Go o Mojo? (golandia 904), corrida 1: el cómputo de la serie, sin servidor web

> Corrida del 2026-11-23 12:26, con el candado `_logs/banco.lock` tomado. Regenerable con `python recursos/pruebas-go-vs-mojo/banco.py 1`. ⛔ Ninguna cifra de blog: todo salió de aquí.

> Sin servidor web: Mojo 1.2 estable no trae TCP en su librería estándar (SALIDA-demos.md). Se miden las MISMAS dos funciones de /suma y /trabajo como programa, con el mismo arnés en los tres.

- go: `go version go1.27.2 linux/amd64`
- mojo: `Mojo 1.2.0 (2016d0cd)`
- python: `Python 3.13.17`
- Docker `29.6.2`. Cada contenedor con 2 CPU y 1g.

## 1 · Construir desde cero (`docker build --no-cache` de la etapa que compila, dependencias incluidas)

| lenguaje | qué se cronometra | segundos |
|---|---|---|
| go | go build | 2.8 |
| mojo | apt install gcc + pip install mojo==1.2.0 + mojo build | 43.2 |
| python | (no compila: se copia el .py) | — |

## 2 · Lo que despliegas

| imagen | base | tamaño |
|---|---|---|
| go | scratch + binario estático | 0.7 MiB |
| mojo | debian:trixie-slim + binario + 3 librerías de Modular | 29.3 MiB |
| python | python:3.13.17-slim + computo.py | 41.1 MiB |

El binario: Go, 1.48 MiB (estático); Mojo, 0.059 MiB, más 1.96 MiB de librerías de Modular que tiene que llevar al lado (libKGENCompilerRTShared, libAsyncRTRuntimeGlobals, libMSupportGlobals).

## Estado de la máquina antes de medir

- /proc/loadavg (desde un contenedor alpine, VM de Docker con 24 hilos): `1.83 1.50 1.62 4/546 6`
- contenedores ajenos en marcha: ninguno

## 0 · ¿Hacen lo mismo? (antes de medir)

- go: total 9990000, primeros ids [0, 1000, 2000, 3000, 4000]…, suma 99863963591 → igual que Go
- mojo: total 9990000, primeros ids [0, 1000, 2000, 3000, 4000]…, suma 99863963591 → igual que Go
- python: total 9990000, primeros ids [0, 1000, 2000, 3000, 4000]…, suma 99863963591 → igual que Go

## 3 · Arranque del proceso (sin argumentos: arranca y termina), mediana de 30, dentro del contenedor

| lenguaje | mediana | p90 |
|---|---|---|
| go | 1.15 ms | 1.38 ms |
| mojo | 6.44 ms | 7.23 ms |
| python | 7.46 ms | 8.20 ms |

## 4 · Cómputo puro: el reloj del propio proceso

50 llamadas de calentamiento y 300 medidas de cada función, de una en una. /suma?n=200000 y /trabajo?n=20000, las mismas del servidor de Go. Memoria máxima del proceso con GNU time (`%M`) durante las 700 llamadas.

| lenguaje | suma mediana | p99 | trabajo mediana | p99 | memoria máx. | tiempo total |
|---|---|---|---|---|---|---|
| go | 0.122 ms | 0.166 ms | 3.274 ms | 4.632 ms | 6.9 MiB | 1.21 s |
| mojo | 0.122 ms | 0.281 ms | 1.684 ms | 2.315 ms | 12.5 MiB | 0.66 s |
| python | 8.136 ms | 8.746 ms | 8.921 ms | 10.872 ms | 14.7 MiB | 6.06 s |

## 5 · `mojo run` contra el binario: lo que tarda `computo una` de punta a punta (mediana de 5)

`mojo run` compila el archivo cada vez antes de ejecutarlo; el binario de `mojo build` ya está compilado.

| cómo | mediana |
|---|---|
| mojo run src/computo.mojo una | 528 ms |
| binario de mojo build: /computo una | 10 ms |
| binario de go build: /computo una | 5 ms |

Al terminar: /proc/loadavg `1.72 1.49 1.61 1/547 5`; ajenos: ninguno

Vigilancia de renders (remotion / fab.py producir) cada 10 s durante toda la corrida: ninguno vivo (orquestar_902_905.py, 2026-11-23 12:27).
