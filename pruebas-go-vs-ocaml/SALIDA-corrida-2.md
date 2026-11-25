# SALIDA — ¿Go o OCaml 5? (golandia 903), corrida 2, con el banco de la serie

> Corrida del 2026-11-23 12:44, con el candado `_logs/banco.lock` tomado. Regenerable con `python investigacion/pruebas_go_vs_x_2026_11_23.py ocaml 2` (mismo banco que 879 y 865). ⛔ Ninguna cifra de blog: todo salió de aquí.

- go: `go version go1.27.2 linux/amd64` (net/http, solo librería estándar)
- ocaml: `5.5.2 · cohttp-eio 6.3.0 · eio        1.7`
- Docker `29.6.2`. Servidores con 2 CPU y 1g; `hey -c 16` con 4 CPU desde otro contenedor.

## 1 · Construir la imagen desde cero (`docker build --no-cache`, dependencias incluidas)

| lenguaje | qué se cronometra | segundos |
|---|---|---|
| go | go build | 5.4 |
| ocaml | opam install dune eio_main cohttp-eio mtime + dune build --profile release | 51.2 |

## 2 · Lo que despliegas

| imagen | base | tamaño |
|---|---|---|
| go | scratch + binario estático | 2.4 MiB |
| ocaml | debian:trixie-slim + binario de ocamlopt (OCaml 5.5.2, cohttp-eio 6.3.0) | 31.6 MiB |

El artefacto que se copia a la imagen: Go, 5.6 MiB; ocaml, 8.98 MiB.

## Estado de la máquina antes de medir

- /proc/loadavg (desde un contenedor alpine, VM de Docker con 24 hilos): `2.24 1.31 1.24 4/518 6`
- contenedores ajenos en marcha: ninguno

## 0 · ¿Hacen lo mismo? (antes de medir)

- go: total 9990000, primeros ids [0, 1000, 2000, 3000, 4000]…, suma 99863963591 → igual que Go
- ocaml (1 dominio, por defecto): total 9990000, primeros ids [0, 1000, 2000, 3000, 4000]…, suma 99863963591 → igual que Go
- ocaml, 2 dominios: total 9990000, primeros ids [0, 1000, 2000, 3000, 4000]…, suma 99863963591 → igual que Go
- ocaml, 24 dominios: total 9990000, primeros ids [0, 1000, 2000, 3000, 4000]…, suma 99863963591 → igual que Go
- ocaml, 2 dominios + io_uring: total 9990000, primeros ids [0, 1000, 2000, 3000, 4000]…, suma 99863963591 → igual que Go

## 3 · Arranque: contenedor nuevo → primera respuesta 200 (mediana de 8) y memoria en reposo

| variante | 1ª respuesta (docker incluido) | «listo» (su propio reloj) | memoria en reposo |
|---|---|---|---|
| go | 188 ms | 0.0 ms | 2.6 MiB |
| ocaml (1 dominio, por defecto) | 201 ms | 1.7 ms | 7.6 MiB |
| ocaml, 2 dominios | 193 ms | 1.7 ms | 11.8 MiB |
| ocaml, 24 dominios | 186 ms | 1.8 ms | 105.8 MiB |
| ocaml, 2 dominios + io_uring | 200 ms | 0.1 ms | 3.9 MiB |

## 4 · Carga con memoria: /trabajo (objetos, mapa, orden, JSON), a tope

`hey -c 16` desde otro contenedor; servidor con 2 CPU y 1g. «En frío»: los primeros 10 s tras arrancar. «Templado»: los 20 s siguientes.

| variante | peticiones/s en frío | peticiones/s templado | p50 | p99 | memoria máx. bajo carga |
|---|---|---|---|---|---|
| go | 481 | 480 | 26.3 ms | 124.2 ms | 37 MiB |
| ocaml (1 dominio, por defecto) | 161 | 161 | 98.8 ms | 104.5 ms | 16 MiB |
| ocaml, 2 dominios | 315 | 317 | 50.4 ms | 53.4 ms | 38 MiB |
| ocaml, 24 dominios | 125 | 159 | 96.7 ms | 282.4 ms | 271 MiB |
| ocaml, 2 dominios + io_uring | 317 | 319 | 45.1 ms | 80.8 ms | 31 MiB |

## 5 · Carga de CPU pura: /suma, a tope

`hey -c 16` desde otro contenedor; servidor con 2 CPU y 1g. «En frío»: los primeros 10 s tras arrancar. «Templado»: los 20 s siguientes.

| variante | peticiones/s en frío | peticiones/s templado | p50 | p99 | memoria máx. bajo carga |
|---|---|---|---|---|---|
| go | 14939 | 14968 | 1.0 ms | 2.6 ms | 11 MiB |
| ocaml (1 dominio, por defecto) | 5135 | 5125 | 3.1 ms | 3.4 ms | 12 MiB |
| ocaml, 2 dominios | 9960 | 9925 | 2.0 ms | 2.4 ms | 21 MiB |
| ocaml, 24 dominios | 5288 | 7430 | 0.5 ms | 71.1 ms | 143 MiB |
| ocaml, 2 dominios + io_uring | 9994 | 9690 | 2.1 ms | 4.1 ms | 13 MiB |

## 6 · Latencia de una petición sola: /trabajo

Un solo cliente a 5 peticiones/s (`hey -c 1 -q 5`), 20 s, tras 10 s de calentamiento a tope. Cada petición llega sola: no hay cola, es lo que tarda el servidor en atender una.

| variante | peticiones/s logradas | p50 | p99 |
|---|---|---|---|
| go | 5 | 6.80 ms | 11.20 ms |
| ocaml (1 dominio, por defecto) | 5 | 8.50 ms | 13.90 ms |
| ocaml, 2 dominios | 5 | 9.50 ms | 13.40 ms |
| ocaml, 24 dominios | 5 | 8.10 ms | 16.40 ms |
| ocaml, 2 dominios + io_uring | 5 | 6.70 ms | 9.00 ms |

## 7 · Latencia de una petición sola: /suma

Un solo cliente a 5 peticiones/s (`hey -c 1 -q 5`), 20 s, tras 10 s de calentamiento a tope. Cada petición llega sola: no hay cola, es lo que tarda el servidor en atender una.

| variante | peticiones/s logradas | p50 | p99 |
|---|---|---|---|
| go | 5 | 0.50 ms | 1.40 ms |
| ocaml (1 dominio, por defecto) | 5 | 0.60 ms | 2.50 ms |
| ocaml, 2 dominios | 5 | 0.60 ms | 1.40 ms |
| ocaml, 24 dominios | 5 | 0.70 ms | 1.90 ms |
| ocaml, 2 dominios + io_uring | 5 | 0.60 ms | 2.20 ms |

## 8 · Cómputo puro: el reloj del propio proceso, sin red

50 peticiones de calentamiento y 300 medidas, de una en una; el tiempo es el campo «ns» que mide cada servidor alrededor de su propio cálculo.

| variante | /suma?n=200000 mediana | p99 | /trabajo?n=20000 mediana | p99 |
|---|---|---|---|---|
| go | 0.122 ms | 0.183 ms | 3.349 ms | 3.896 ms |
| ocaml (1 dominio, por defecto) | 0.180 ms | 0.234 ms | 6.297 ms | 6.933 ms |
| ocaml, 2 dominios | 0.180 ms | 0.337 ms | 5.878 ms | 7.113 ms |
| ocaml, 24 dominios | 0.180 ms | 0.257 ms | 6.040 ms | 10.539 ms |
| ocaml, 2 dominios + io_uring | 0.180 ms | 0.217 ms | 5.215 ms | 6.632 ms |

Al terminar: /proc/loadavg `0.64 1.00 1.17 3/527 6`; ajenos: ninguno

Vigilancia de renders (remotion / fab.py producir) cada 10 s durante toda la corrida: ninguno vivo (orquestar_902_905.py, 2026-11-23 12:59).
